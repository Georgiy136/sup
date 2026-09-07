package createoperations

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createoperationmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_operations/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerCreateOperation"

	statusIDForGetPerformer = "EOO"
)

type TicketHandlerCreateOperation struct {
	prodTypesApi prodTypesApi
	repo         services.HandlerTicketsRepo
}

func NewTicketHandlerCreateOperation(repo services.HandlerTicketsRepo, prodTypesApi prodTypesApi) *TicketHandlerCreateOperation {
	return &TicketHandlerCreateOperation{
		prodTypesApi: prodTypesApi,
		repo:         repo,
	}
}

func (h *TicketHandlerCreateOperation) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createoperationmodels.ExtTicketInfoCreationOperation

	err := handlerutils.DecodeMapToStructureWithoutErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	bodyRequestForProdtypesAddNew := prepareRequestForProdtypesAddNew(ext)

	var employeeID int64

	for i := range ticketInfo.StatusFields {
		if ticketInfo.StatusFields[i].StatusID == statusIDForGetPerformer {
			employeeID = ticketInfo.StatusFields[i].PerformEmployeeID
		}
	}

	if employeeID == 0 {
		return fmt.Errorf("[%s] can't find employee id for this ticket", handlerName)
	}

	err = h.prodTypesApi.ProdtypesAddNew(ctx, bodyRequestForProdtypesAddNew, employeeID)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("[%s] can't add new prodtype err: %w", handlerName, err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

type prodTypesApi interface {
	ProdtypesAddNew(ctx context.Context, body createoperationmodels.RequestDataForProdtypesAddNew, employeeID int64) (err error)
}
