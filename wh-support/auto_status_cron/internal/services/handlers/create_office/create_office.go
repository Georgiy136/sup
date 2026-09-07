package createoffice

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createofficemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_office/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	statusIDCreateOffice = "EXP"
)

type TicketHandlerCreateOffice struct {
	repo            services.HandlerTicketsRepo
	createOfficeApi createOfficeApi
}

func NewTicketHandlerCreateOffice(repo services.HandlerTicketsRepo, createOfficeApi createOfficeApi) *TicketHandlerCreateOffice {
	return &TicketHandlerCreateOffice{
		repo:            repo,
		createOfficeApi: createOfficeApi,
	}
}

func (h *TicketHandlerCreateOffice) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if ticketInfo.StatusID == statusIDCreateOffice {
		if err := h.createOffice(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't create office: %w", err)
		}
		return nil
	}
	return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
}

func (h *TicketHandlerCreateOffice) createOffice(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createofficemodels.ExtTicketInfoForCreateOffice

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	requestBodyForCreateOffice := convertExtCreateOfficeRequestForCreateOffice(ext)

	requestBodyForCreateOffice.EmployeeId = ticketInfo.CreateEmployeeID

	err = h.createOfficeApi.CreateOfficeOnSprut(ctx, requestBodyForCreateOffice)
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
		return fmt.Errorf("can't create office on sprut err: %w", err)
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

type createOfficeApi interface {
	CreateOfficeOnSprut(ctx context.Context, body createofficemodels.RequestForCreateOffice) (err error)
}
