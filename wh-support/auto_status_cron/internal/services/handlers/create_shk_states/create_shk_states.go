package createshkstates

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createshkstatesmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_shk_states/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

type TicketHandlerCreateShkStates struct {
	repo   services.HandlerTicketsRepo
	shkApi shkApi
}

func NewTicketHandlerCreateShkStates(repo services.HandlerTicketsRepo, client shkApi) *TicketHandlerCreateShkStates {
	return &TicketHandlerCreateShkStates{
		repo:   repo,
		shkApi: client,
	}
}

func (s *TicketHandlerCreateShkStates) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	const defaultLocLang = "ru"

	var ext createshkstatesmodels.ExtTicketInfoForCreateShkStates

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	err = s.shkApi.ShkStatesCreate(ctx, createshkstatesmodels.RequestForCreateShkStates{
		StateID:    ext.StateID,
		StateDescr: ext.StateName,
		ProcessID:  ext.ProcessID.ID,
		LocLang:    defaultLocLang,
	}, ticketInfo.CreateEmployeeID)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := s.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d about create shk states: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't create shk states: %w", err)
	}

	err = s.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := s.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

type shkApi interface {
	ShkStatesCreate(ctx context.Context, body createshkstatesmodels.RequestForCreateShkStates, employeeID int64) (err error)
}
