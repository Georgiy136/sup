package createstreets

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createstreetsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_streets/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerCreateStreets"
)

type TicketHandlerCreateStreets struct {
	createStreetApi createStreetApi
	repo            services.HandlerTicketsRepo
}

func NewTicketHandlerCreateStreets(repo services.HandlerTicketsRepo, createStreetApi createStreetApi) *TicketHandlerCreateStreets {
	return &TicketHandlerCreateStreets{
		repo:            repo,
		createStreetApi: createStreetApi,
	}
}

func (c *TicketHandlerCreateStreets) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstreetsmodels.ExtTicketInfoForCreateStreets

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	stages := make([]int64, len(ext.Stages))
	for i := range ext.Stages {
		stages[i] = ext.Stages[i].ID
	}

	if err = c.createStreetApi.CreateStreet(ctx, convertExtInfoToRequestForCreateStreet(ext, stages, ticketInfo.CreateEmployeeID)); err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := c.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("[%s] can't add create street err: %w", handlerName, err)
	}

	if err = c.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

type createStreetApi interface {
	CreateStreet(ctx context.Context, body createstreetsmodels.RequestForCreateStreet) error
}
