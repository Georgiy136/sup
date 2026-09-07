package createtarestate

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createtarestatemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_tare_state/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

type TicketHandlerCreateTareState struct {
	repo                services.HandlerTicketsRepo
	tareStateApiCreator tareStateApiCreator
}

func NewTicketHandlerCreateTareState(repo services.HandlerTicketsRepo, createTareStateApi tareStateApiCreator) *TicketHandlerCreateTareState {
	return &TicketHandlerCreateTareState{
		repo:                repo,
		tareStateApiCreator: createTareStateApi,
	}
}

func (h *TicketHandlerCreateTareState) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if err := h.createTareState(ctx, ticketInfo); err != nil {
		return fmt.Errorf("can't create tare state: %w", err)
	}
	return nil
}

func (h *TicketHandlerCreateTareState) createTareState(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	const defaultLocLang = "ru"

	var ext createtarestatemodels.ExtTicketInfoForCreateTareState

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	err = h.tareStateApiCreator.CreateTareState(ctx, createtarestatemodels.RequestForTareStateCreate{
		StateID:   ext.StateID,
		StateDesc: ext.StateDescr,
		LocLang:   defaultLocLang,
	}, ticketInfo.CreateEmployeeID)
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
		return fmt.Errorf("can't create tare state err: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}
	return nil
}

type tareStateApiCreator interface {
	CreateTareState(ctx context.Context, body createtarestatemodels.RequestForTareStateCreate, employeeID int64) error
}
