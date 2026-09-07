package closewh

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	closewhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh/models"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/sirupsen/logrus"
)

const (
	statusIdCloseWh                    = "DEL"
	statusIdCheckUndeletedStoragePlace = "CMH"

	defaultRejectMsgUndeletedStoragePlace = "На блоке %d имеются неудаленные МХ"
)

type TicketHandlerCloseWh struct {
	whApi whApi
	repo  services.HandlerTicketsRepo
}

func NewTicketHandlerCloseWh(repo services.HandlerTicketsRepo, whApi whApi) *TicketHandlerCloseWh {
	return &TicketHandlerCloseWh{
		repo:  repo,
		whApi: whApi,
	}
}

func (h *TicketHandlerCloseWh) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdCloseWh:
		if err := h.closeWh(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't close wh: %w", err)
		}
	case statusIdCheckUndeletedStoragePlace:
		if err := h.checkUndeletedStoragePlaces(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't check undeleted storage places: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerCloseWh) checkUndeletedStoragePlaces(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodels.ExtTicketInfoForCheckUndeletedStoragePlaces
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}
	bodyRequestForGetStoragePlaces := prepareRequestForGetStoragePlacesByWh(ext)

	storagePlaces, err := h.whApi.GetStoragePlacesByWh(ctx, bodyRequestForGetStoragePlaces)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("can't get storage places by Wh, err: %w", err)
	}

	if len(storagePlaces) != 0 {
		locMsgs := localization.New(
			support_err_keys.KeyErrorBlockHasUndeletedStorage,
			map[string]string{
				"wh_id": fmt.Sprintf("%d", ext.WhId.Id),
			},
		)

		rejectValues := map[string]any{
			"values": storagePlaces,
		}

		rejectErr := h.repo.RejectTicketWithValues(ctx,
			ticketInfo.TicketID,
			fmt.Sprintf(defaultRejectMsgUndeletedStoragePlace, ext.WhId.Id),
			locMsgs,
			rejectValues,
		)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		return nil
	}

	if err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCloseWh) closeWh(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodels.ExtTicketInfoForCloseWh
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	bodyRequestForCloseWh := prepareRequestForCloseWh(ext)
	bodyRequestForCloseWh.EmployeeId = ticketInfo.CreateEmployeeID
	if err := h.whApi.CloseWh(ctx, bodyRequestForCloseWh); err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("can't close wh err: %w", err)
	}

	if err := h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}
		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

type whApi interface {
	CloseWh(ctx context.Context, body closewhmodelsnew.RequestForCloseWh) error
	GetStoragePlacesByWh(ctx context.Context, body closewhmodelsnew.RequestGetStoragePlacesByWh) ([]int64, error)
}
