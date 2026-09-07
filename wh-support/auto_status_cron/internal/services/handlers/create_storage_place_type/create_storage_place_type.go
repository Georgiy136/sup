package createstorageplacetype

import (
	"context"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createstorageplacetypemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_type/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

type TicketHandlerCreateStoragePlaceType struct {
	repo                       services.HandlerTicketsRepo
	storagePlaceTypeApiCreator storagePlaceTypeApiCreator
}

func NewTicketHandlerCreateStoragePlaceType(repo services.HandlerTicketsRepo, createStoragePlaceTypeApi storagePlaceTypeApiCreator) *TicketHandlerCreateStoragePlaceType {
	return &TicketHandlerCreateStoragePlaceType{
		repo:                       repo,
		storagePlaceTypeApiCreator: createStoragePlaceTypeApi,
	}
}

func (h *TicketHandlerCreateStoragePlaceType) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if err := h.createStoragePlaceType(ctx, ticketInfo); err != nil {
		return fmt.Errorf("can't create storage place type: %w", err)
	}
	return nil
}

func (h *TicketHandlerCreateStoragePlaceType) createStoragePlaceType(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	const defaultLocLang = "ru"

	var ext createstorageplacetypemodels.ExtTicketInfoForCreateStoragePlaceType

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	resp, err := h.storagePlaceTypeApiCreator.CreateStoragePlaceType(ctx, createstorageplacetypemodels.RequestForCreateStoragePlaceType{
		PlaceTypeName: ext.PlaceTypeName,
		EmployeeID:    ticketInfo.CreateEmployeeID,
		PlaceTypeID:   nil,
		StickerType:   ext.StickerType,
		StickerPrefix: ext.StickerPrefix.ID,
		LocLang:       defaultLocLang,
		IsDel:         false,
	})

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
		return fmt.Errorf("can't create storage place type err: %w", err)
	}

	addedInfoStoragePlaceType := createstorageplacetypemodels.AddedPerformInfoStoragePlaceType{
		PlaceTypeID: resp.PlaceTypeID,
	}
	rawAddedInfoStoragePlaceType, err := jsoniter.Marshal(addedInfoStoragePlaceType)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoStoragePlaceType: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoStoragePlaceType)

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

type storagePlaceTypeApiCreator interface {
	CreateStoragePlaceType(ctx context.Context, body createstorageplacetypemodels.RequestForCreateStoragePlaceType) (*createstorageplacetypemodels.StoragePlaceType, error)
}
