package assemblysheetterminate

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	assemblysheetterminatemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/assembly_sheet_terminate/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const handlerName = "TicketHandlerAssemblySheetTerminate"

type TicketHandlerAssemblySheetTerminate struct {
	repo                      services.HandlerTicketsRepo
	assemblySheetTerminateApi assemblySheetTerminateApi
}

func NewTicketHandlerAssemblySheetTerminate(
	repo services.HandlerTicketsRepo,
	assemblySheetTerminateApi assemblySheetTerminateApi,
) *TicketHandlerAssemblySheetTerminate {
	return &TicketHandlerAssemblySheetTerminate{
		repo:                      repo,
		assemblySheetTerminateApi: assemblySheetTerminateApi,
	}
}

func (h *TicketHandlerAssemblySheetTerminate) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if err := h.terminateAssemblySheet(ctx, ticketInfo); err != nil {
		return fmt.Errorf("[%s] can't terminate assembly sheet: %w", handlerName, err)
	}

	return nil
}

func (h *TicketHandlerAssemblySheetTerminate) terminateAssemblySheet(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext assemblysheetterminatemodels.ExtTicketInfoForAssemblySheetTerminate
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}

	err := h.assemblySheetTerminateApi.AssemblySheetTerminate(ctx, assemblysheetterminatemodels.RequestDataForAssemblySheetTerminate{
		OfficeID:         ext.OfficeID.ID,
		WhID:             ext.WhID.ID,
		EmployeeID:       ext.EmployeeID.ID,
		AsmsheetID:       ext.AsmsheetID,
		CreateEmployeeID: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("[%s] can't terminate assembly sheet err: %w", handlerName, err)
	}

	if err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}
		return fmt.Errorf("[%s] can't perform ticket %d: %w", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

type assemblySheetTerminateApi interface {
	AssemblySheetTerminate(ctx context.Context, body assemblysheetterminatemodels.RequestDataForAssemblySheetTerminate) error
}
