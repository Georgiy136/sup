package excludetare

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	excludetaremodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_tare/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	statusIDExcludeTare              = "AUT"
	statusIDExcludeTareOnWriteoffApi = "OUT"
)

type TicketHandlerExcludeTare struct {
	repo        services.HandlerTicketsRepo
	tareApi     tareApi
	writeoffApi writeoffApi
}

func NewTicketHandlerExcludeTare(repo services.HandlerTicketsRepo, tareApi tareApi, writeoffApi writeoffApi) *TicketHandlerExcludeTare {
	return &TicketHandlerExcludeTare{
		repo:        repo,
		tareApi:     tareApi,
		writeoffApi: writeoffApi,
	}
}

func (h *TicketHandlerExcludeTare) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIDExcludeTare:
		err := h.excludeTare(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't exclude tare: %w", err)
		}
	case statusIDExcludeTareOnWriteoffApi:
		err := h.excludeTareOnWriteoffApi(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't exclude tare on writeoff api: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerExcludeTare) excludeTare(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext excludetaremodels.ExtTicketInfoForExcludeTare

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	reqBody := prepareRequestDataForExcludeTare(ext)

	err = h.tareApi.ExcludeTare(ctx, reqBody, ticketInfo.CreateEmployeeID)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't exclude tare err: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerExcludeTare) excludeTareOnWriteoffApi(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext excludetaremodels.ExtTicketInfoForExcludeTare

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	reqBody := prepareRequestDataForExcludeTareOnWriteoffApi(ext, ticketInfo.CreateEmployeeID)

	err = h.writeoffApi.ExcludeTareOnWriteoff(ctx, reqBody)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't exclude tare on writeoff api err: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

type tareApi interface {
	ExcludeTare(ctx context.Context, body excludetaremodels.RequestBodyForExcludeTare, employeeID int64) error
}

type writeoffApi interface {
	ExcludeTareOnWriteoff(ctx context.Context, body []excludetaremodels.RequestBodyForExcludeTareOnWriteoffApi) error
}
