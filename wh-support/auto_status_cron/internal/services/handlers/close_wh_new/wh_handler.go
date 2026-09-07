package closewh

import (
	"context"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/sirupsen/logrus"
)

func (t *TicketHandlerCloseWhNew) closeWh(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForCloseWh
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode ext: %w", err)
	}

	body := prepareRequestForCloseWh(ext, ticketInfo.CreateEmployeeID)

	if err := t.whApi.CloseWh(ctx, body); err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			rawErrValues, err := jsoniter.Marshal(models.PerformErrValue{Comment: errWithMsg.Msg})
			if err != nil {
				return fmt.Errorf("can't marshal performErr for ticket %d: %w", ticketInfo.TicketID, err)
			}
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if err = t.repo.PerformTicketWithScenarioAndErrComment(ctx, ticketInfo.TicketID, nil, resolveCloseWhScenario(closewhmodelsnew.ActionFailed), rawErrValues, locMsgs); err != nil {
				return fmt.Errorf("can't perform ticket %d with err comment: %w", ticketInfo.TicketID, err)
			}
			logrus.Infof("ticket %d performed with err comment: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("close wh err: %w", err)
	}

	if err := t.repo.PerformTicketWithScenario(ctx, ticketInfo.TicketID, nil, resolveCloseWhScenario(closewhmodelsnew.ActionCompleted)); err != nil {
		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}
	return nil
}

func (t *TicketHandlerCloseWhNew) deactivateWh(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForDeactivateWh
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	body := prepareRequestForDeactivateWh(ext, ticketInfo.CreateEmployeeID)

	if err := t.whApi.DeactivateWh(ctx, body); err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("can't deactivate wh err: %w", err)
	}

	if err := t.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}
	return nil
}
