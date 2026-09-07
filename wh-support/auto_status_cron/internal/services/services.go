package services

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	serviceerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	overrideCategoriesConfigKey = "override_categories"
)

type ticketHandler interface {
	Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error
}

type batchTicketHandler interface {
	ProcessBatch(ctx context.Context, tickets []models.TicketCommonInfo) map[int64]error
}

type AutoStatusService struct {
	ticketHandlerRegistry      map[int64]ticketHandler
	batchTicketHandlerRegistry map[int64]batchTicketHandler
	repo                       ServiceTicketsRepo

	IsOverrideCategory bool
	EnabledCategories  map[int64]struct{}
}

func NewAutoStatusService(repo ServiceTicketsRepo) *AutoStatusService {
	return &AutoStatusService{
		ticketHandlerRegistry:      make(map[int64]ticketHandler),
		batchTicketHandlerRegistry: make(map[int64]batchTicketHandler),
		EnabledCategories:          make(map[int64]struct{}),
		repo:                       repo,
	}
}

func (a *AutoStatusService) Configure(ctx context.Context, config configs.Config) {
	ok, rawConfig := config.GetByServiceKey(overrideCategoriesConfigKey)
	if !ok {
		logrus.Infof("override_categories not found")
		return
	}

	var overrideCategoriesConfig models.OverrideCategoriesConfig
	if err := jsoniter.Unmarshal(rawConfig, &overrideCategoriesConfig); err != nil {
		logrus.Errorf("can't unmarshal override categories config: %v", err)
	}

	for _, category := range overrideCategoriesConfig.EnabledCategories {
		a.EnabledCategories[category] = struct{}{}
	}

	a.IsOverrideCategory = overrideCategoriesConfig.IsOverrideCategory

	if a.IsOverrideCategory {
		logrus.Infof("override enabled categories: %v", overrideCategoriesConfig)
	}
}

func (a *AutoStatusService) ProcessTicketsOnAutoStatus(ctx context.Context) error {
	ticketsData, err := a.repo.GetTicketsWithAutoStatus(ctx)
	if err != nil {
		return fmt.Errorf("can't get tickets with auto status: %w", err)
	}

	if ticketsData == nil {
		logrus.Info("no tickets found...")
		return nil
	}

	ticketsByCategory := a.groupTicketsByCategory(ticketsData.Tickets)

	isSuccess := true

	for categoryID, tickets := range ticketsByCategory {
		if _, exist := a.EnabledCategories[categoryID]; a.IsOverrideCategory && !exist {
			logrus.Infof("handle category %d skip", categoryID)
			continue
		}

		// Проверяем, есть ли batch-обработчик для данной категории
		if batchHandler, ok := a.batchTicketHandlerRegistry[categoryID]; ok {

			// Логируем результат по каждой заявке отдельно
			errorsMap := a.processBatch(ctx, tickets, batchHandler)
			for _, ticket := range tickets {
				if ticketErr, hasErr := errorsMap[ticket.TicketID]; hasErr && ticketErr != nil {
					logrus.Errorf("can't handle ticket %d, handler for category_id=%d err: %v", ticket.TicketID, categoryID, ticketErr)
					isSuccess = false
				} else {
					logrus.Debugf("ticket № %d processed...", ticket.TicketID)
				}
			}
			continue
		}

		handler, ok := a.ticketHandlerRegistry[categoryID]
		if !ok {
			for _, ticket := range tickets {
				logrus.Errorf("can't handle ticket %d, handler for category_id=%d not found", ticket.TicketID, categoryID)
			}
			continue
		}

		for _, ticket := range tickets {
			if err = a.processTicket(ctx, ticket, handler); err != nil {
				logrus.Errorf("can't handle ticket %d, status %s handler for category_id=%d err: %v", ticket.TicketID, ticket.StatusID, categoryID, err)
				isSuccess = false
				continue
			}
			logrus.Debugf("ticket № %d processed...", ticket.TicketID)
		}
	}

	if !isSuccess {
		return fmt.Errorf("can't process ticket: %w", serviceerrors.ErrDuringTicketsProcess)
	}

	return nil
}

func (a *AutoStatusService) groupTicketsByCategory(tickets []models.TicketCommonInfo) map[int64][]models.TicketCommonInfo {
	result := make(map[int64][]models.TicketCommonInfo, len(tickets))

	for _, ticket := range tickets {
		result[ticket.CategoryID] = append(result[ticket.CategoryID], ticket)
	}

	return result
}

func (a *AutoStatusService) processBatch(ctx context.Context, tickets []models.TicketCommonInfo, handler batchTicketHandler) map[int64]error {
	tx, traceCtx := sentry.StartBatchTransaction(ctx, tickets)

	span, traceCtx := sentry.StartTicketHandlerSpan(traceCtx)
	ticketErrors := handler.ProcessBatch(traceCtx, tickets)

	var batchErr error
	for _, ticket := range tickets {
		if ticketErr := ticketErrors[ticket.TicketID]; ticketErr != nil {
			batchErr = errors.Join(batchErr, ticketErr)
			sentry.CaptureTicketException(traceCtx, ticket, ticketErr)
		}
	}
	sentry.FinishTicketHandlerSpan(span, batchErr)
	if batchErr != nil {
		sentry.MarkTransactionFailed(tx)
		tx.Finish()
	}
	return ticketErrors
}

func (a *AutoStatusService) processTicket(ctx context.Context, ticket models.TicketCommonInfo, handler ticketHandler) (err error) {
	tx, traceCtx := sentry.StartTicketTransaction(ctx, ticket)
	defer func() {
		if err != nil {
			sentry.MarkTransactionFailed(tx)
			sentry.CaptureTicketException(traceCtx, ticket, err)
			tx.Finish()
		}
	}()

	span, traceCtx := sentry.StartTicketHandlerSpan(traceCtx)
	defer func() {
		sentry.FinishTicketHandlerSpan(span, err)
	}()
	return handler.Process(traceCtx, ticket)
}

func (a *AutoStatusService) RegisterTicketHandler(categoryID int64, handler ticketHandler) {
	if _, ok := a.ticketHandlerRegistry[categoryID]; ok {
		logrus.Panic("category id already registered")
	}

	a.ticketHandlerRegistry[categoryID] = handler
}

// RegisterBatchTicketHandler регистрирует batch-обработчик для категории
func (a *AutoStatusService) RegisterBatchTicketHandler(categoryID int64, handler batchTicketHandler) {
	if _, ok := a.batchTicketHandlerRegistry[categoryID]; ok {
		logrus.Panic("batch handler for category id already registered")
	}

	a.batchTicketHandlerRegistry[categoryID] = handler
}

type ServiceTicketsRepo interface {
	GetTicketsWithAutoStatus(ctx context.Context) (*models.DataTickets, error)
}
