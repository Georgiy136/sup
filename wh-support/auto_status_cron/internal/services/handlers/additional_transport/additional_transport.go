package additionaltransport

import (
	"context"
	"errors"
	"fmt"
	"time"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	additionaltransportmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/additional_transport/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	statusIDCreateLogisticTicketEX1 = "EX1"
	statusIDCreateLogisticTicketEA1 = "EA1"
	statusIDWaitLogisticTicketEX2   = "EX2"
	statusIDWaitLogisticTicketEA2   = "EA2"

	formatErrMsgForCanceledLogisticTicket = "Заявка отменена в системе логистики. Причина: %s"
	errMsgForErrWhileCancelLogisticTicket = "Запрос на отклонение заявки отменен. В системе логистики произошла ошибка. Обратитесь в тех. поддержку"
	errMsgForErrForCreateLogisticTicket   = "Ошибка при создании заявки в системе логистики. Обратитесь в тех. поддержку."
	errMsgForExpiresCargoDate             = "Ошибка при создании заявки в системе логистики. Дата постановки ТС просрочена."
)

type AdditionalTransport struct {
	repo                  services.HandlerTicketsRepo
	logisticCargoResolver externalLogisticCargoResolver
}

func NewTicketHandlerAdditionalTransport(repo services.HandlerTicketsRepo, logisticCargoResolver externalLogisticCargoResolver) *AdditionalTransport {
	return &AdditionalTransport{
		repo:                  repo,
		logisticCargoResolver: logisticCargoResolver,
	}
}

var additionalTransportStatuses = []string{
	statusIDCreateLogisticTicketEX1,
	statusIDCreateLogisticTicketEA1,
	statusIDWaitLogisticTicketEX2,
	statusIDWaitLogisticTicketEA2,
}

func (a *AdditionalTransport) ProcessBatch(ctx context.Context, tickets []models.TicketCommonInfo) map[int64]error {
	groupsTickets, ticketsErrMap := handlerutils.GroupTicketsByStatuses(tickets, additionalTransportStatuses)
	if len(ticketsErrMap) != 0 {
		logByTicketsErrMap(ticketsErrMap)
	}

	for status, ticketsGroup := range groupsTickets {
		switch status {
		case statusIDCreateLogisticTicketEX1, statusIDCreateLogisticTicketEA1:
			ticketsErrMap = a.createLogisticTicketByGroup(ctx, ticketsGroup)
			if len(ticketsErrMap) != 0 {
				return ticketsErrMap
			}
		case statusIDWaitLogisticTicketEX2, statusIDWaitLogisticTicketEA2:
			filteredTickets, preRejectTicketsErrMap := a.checkAndHandlePreRejectByGroup(ctx, ticketsGroup)

			ticketsErrMap = a.waitLogisticTicketByGroup(ctx, filteredTickets)

			ticketsErrMap = handlerutils.MergeMapsOverwrite(ticketsErrMap, preRejectTicketsErrMap)

			if len(ticketsErrMap) != 0 {
				return ticketsErrMap
			}
		default:
			logrus.Errorf("unknown status for additional transport category: %s", status)
		}
	}

	return nil
}

func logByTicketsErrMap(ticketsErrMap map[int64]error) {
	for _, err := range ticketsErrMap {
		logrus.Error(err)
	}
}

func (a *AdditionalTransport) checkAndHandlePreRejectByGroup(ctx context.Context, tickets []models.TicketCommonInfo) ([]models.TicketCommonInfo, map[int64]error) {
	ticketsErr := make(map[int64]error)

	filteredTickets := make([]models.TicketCommonInfo, 0)

	for _, ticket := range tickets {
		if !ticket.IsPrereject {
			filteredTickets = append(filteredTickets, ticket)
			continue
		}

		err := a.checkAndHandlePreReject(ctx, ticket)
		if err != nil {
			ticketsErr[ticket.TicketID] = fmt.Errorf("can't check and handle prereject ticket %d: %w", ticket.TicketID, err)
		}
	}

	return filteredTickets, ticketsErr
}

func (a *AdditionalTransport) checkAndHandlePreReject(ctx context.Context, ticket models.TicketCommonInfo) error {
	var ext additionaltransportmodels.ExtTicketInfosForWaitLogisticTicket

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticket.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	resp, err := a.logisticCargoResolver.CanRejectInternalTicket(ctx, additionaltransportmodels.RequestForCanRejectInternalTicket{
		CargoId: ext.CargoId,
	})
	if err != nil {
		return fmt.Errorf("can't check can reject internal ticket: %w", err)
	}

	if !resp.Success {
		err := a.repo.CancelPreReject(ctx,
			ticket.TicketID,
			map[string]any{"values": resp.Reason},
		)
		if err != nil {
			return fmt.Errorf("can't cancel prereject ticket: %w", err)
		}

		return nil
	}

	err = a.logisticCargoResolver.RejectLogisticTicket(ctx, additionaltransportmodels.RequestRejectLogisticTicket{
		CargoId: ext.CargoId,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			preRejectErr := a.repo.CancelPreReject(ctx, ticket.TicketID, map[string]any{"values": fmt.Sprintf(errMsgForErrWhileCancelLogisticTicket)})
			if preRejectErr != nil {
				return fmt.Errorf("can't cancel prereject ticket while reject logistic ticket: %w", preRejectErr)
			}

			logrus.Infof("prereject cancel for ticket %d while reject logistic ticket: %v", ticket.TicketID, errWithMsg.Msg)

			return nil
		}

		return fmt.Errorf("can't reject logistic ticket: %w", err)
	}

	err = a.repo.RejectTicketWithEmployee(ctx, ticket.TicketID, ticket.CreateEmployeeID, ticket.PrerejectComment, nil)
	if err != nil {
		return fmt.Errorf("can't reject ticket: %v", err)
	}

	return nil
}

func (a *AdditionalTransport) waitLogisticTicketByGroup(ctx context.Context, tickets []models.TicketCommonInfo) map[int64]error {
	ticketsErr := make(map[int64]error)

	cargoIdToTicketId := make(map[string]int64)
	exts := make([]additionaltransportmodels.ExtTicketInfosForWaitLogisticTicket, 0, len(tickets))

	for _, ticket := range tickets {
		var ext additionaltransportmodels.ExtTicketInfosForWaitLogisticTicket

		err := handlerutils.DecodeMapToStructureWithErrorUnset(ticket.Ext, &ext)
		if err != nil {
			ticketsErr[ticket.TicketID] = fmt.Errorf("can't decode request body by wait logistic ticket: %w", err)
			continue
		}

		exts = append(exts, ext)
		cargoIdToTicketId[ext.CargoId] = ticket.TicketID
	}

	req := additionaltransportmodels.RequestForResolveInternalTicketStateByCargo{
		CargoIds: make([]string, len(exts)),
	}

	for idx, ext := range exts {
		req.CargoIds[idx] = ext.CargoId
	}

	decisions, err := a.logisticCargoResolver.ResolveInternalTicketsStateByCargos(ctx, req)
	if err != nil {
		logrus.Errorf("can't resolve internal tickets state by cragos: %v", err)
		return nil
	}

	for _, decision := range decisions {
		ticketId, ok := cargoIdToTicketId[decision.CargoId]
		if !ok {
			logrus.Errorf("can't find support ticket fot cargo id %s", decision.CargoId)
			continue
		}

		switch decision.Action {
		case InternalTicketActionReject:
			rejectErr := a.repo.RejectTicket(ctx, ticketId, fmt.Sprintf(formatErrMsgForCanceledLogisticTicket, decision.CancelReason), nil)
			if rejectErr != nil {
				ticketsErr[ticketId] = fmt.Errorf(" can't reject ticket %d before canceled logistic ticket: %w", ticketId, rejectErr)
				continue
			}
		case InternalTicketActionComplete:
			rawAddedInfoDriver, err := jsoniter.Marshal(decision.PerformInfo)
			if err != nil {
				ticketsErr[ticketId] = fmt.Errorf("can't marshal addedInfoDriver: %w", err)
				continue
			}

			err = a.repo.PerformTicket(ctx, ticketId, rawAddedInfoDriver)
			if err != nil {
				locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
				rejectErr := a.repo.RejectTicket(ctx, ticketId, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
				if rejectErr != nil {
					logrus.Errorf("can't reject ticket %d while perform ticket on wait logistic ticket: %v", ticketId, rejectErr)
				}

				ticketsErr[ticketId] = fmt.Errorf("can't perform ticket %d on wait logistic ticket: %w", ticketId, err)
				continue
			}
		case InternalTicketActionNone:
			continue
		default:
			logrus.Errorf("unknown internal ticket action: %v", decision.Action)
		}
	}

	return ticketsErr
}

func (a *AdditionalTransport) createLogisticTicketByGroup(ctx context.Context, tickets []models.TicketCommonInfo) map[int64]error {
	ticketsErr := make(map[int64]error)

	for _, ticket := range tickets {
		err := a.createLogisticTicket(ctx, ticket)
		if err != nil {
			ticketsErr[ticket.TicketID] = fmt.Errorf("can't create logistic ticket by ticket %d: %w", ticket.TicketID, err)
		}
	}

	return ticketsErr
}

func (a *AdditionalTransport) createLogisticTicket(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext additionaltransportmodels.ExtTicketInfosForCreateLogisticTicket

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	cargoDateTime, err := time.Parse(time.DateOnly, ext.Date)
	if err != nil {
		return fmt.Errorf("can't convert date to RFC3339 format: %w", err)
	}

	if cargoDateTime.Before(time.Now()) {
		rejectErr := a.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsgForExpiresCargoDate, nil)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errMsgForExpiresCargoDate)

		return nil
	}

	req, err := prepareRequestForLogisticCargo(ticketInfo.TicketID, ext)
	if err != nil {
		return fmt.Errorf("can't prepare request for logistic cargo: %w", err)
	}

	cargoId, err := a.logisticCargoResolver.CreateLogisticCargo(ctx, req)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			rejectErr := a.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsgForErrForCreateLogisticTicket, nil)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject while create logistic cargo: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}

		return fmt.Errorf("can't create logistic ticket: %w", err)
	}

	var addedInfoCargoId additionaltransportmodels.AddedPerformInfoCargoId

	addedInfoCargoId.CargoId = cargoId

	rawAddedInfoCargoId, err := jsoniter.Marshal(addedInfoCargoId)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoCargoId: %w", err)
	}

	err = a.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoCargoId)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := a.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d while perform: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

type externalLogisticCargoResolver interface {
	CreateLogisticCargo(ctx context.Context, req additionaltransportmodels.RequestForCreateLogisticCargo) (string, error)
	RejectLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestRejectLogisticTicket) error

	ResolveInternalTicketsStateByCargos(ctx context.Context, req additionaltransportmodels.RequestForResolveInternalTicketStateByCargo) ([]additionaltransportmodels.ResponseResolveInternalTicketStateByCargo, error)
	CanRejectInternalTicket(ctx context.Context, req additionaltransportmodels.RequestForCanRejectInternalTicket) (additionaltransportmodels.ResponseCanRejectInternalTicket, error)
}
