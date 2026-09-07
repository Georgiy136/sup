package deletestorageplace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	deletestorageplacemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/delete_storage_place/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

const (
	handlerName = "TicketHandlerDeleteStoragePlace"

	errWithStoragePlaces = "%s Список МХ, из-за которых не удалось выполнить заявку, в прикрепленном файле."
)

type TicketHandlerDeleteStoragePlace struct {
	repo                   services.HandlerTicketsRepo
	storagePlaceApiDeleter storagePlaceApiDeleter
}

func NewTicketHandlerDeleteStoragePlace(repo services.HandlerTicketsRepo, storagePlaceApiDeleter storagePlaceApiDeleter) *TicketHandlerDeleteStoragePlace {
	return &TicketHandlerDeleteStoragePlace{
		repo:                   repo,
		storagePlaceApiDeleter: storagePlaceApiDeleter,
	}
}

func (h *TicketHandlerDeleteStoragePlace) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if err := h.deleteStoragePlace(ctx, ticketInfo); err != nil {
		return fmt.Errorf("[%s] can't delete storage place: %w", handlerName, err)
	}
	return nil
}

func (h *TicketHandlerDeleteStoragePlace) deleteStoragePlace(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext deletestorageplacemodels.ExtTicketInfoForDeleteStoragePlace

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}

	placeIDsMap := make(map[int64]struct{}, len(ext.Places))
	for _, place := range ext.Places {
		placeIDsMap[place.PlaceID.Value] = struct{}{}
	}
	placeIDs := maps.Keys(placeIDsMap)

	err = h.storagePlaceApiDeleter.DeleteStoragePlace(ctx, deletestorageplacemodels.RequestDataForDeleteStoragePlace{
		OfficeID:   ext.OfficeID.ID,
		PlaceIDs:   placeIDs,
		EmployeeID: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			var rejectErr error
			if errWithMsg.Detail != "" && strings.HasPrefix(errWithMsg.Detail, "МХ:") {
				comment := fmt.Sprintf(errWithStoragePlaces, errWithMsg.Msg)
				storagePlaces := map[string]any{
					"values": parseStoragePlacesFromDetail(errWithMsg.Detail),
				}

				locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
				locMsgs.Add(support_err_keys.KeyErrorStoragePlaceListAttached, nil)

				rejectErr = h.repo.RejectTicketWithValues(ctx, ticketInfo.TicketID, comment, locMsgs, storagePlaces)
			} else {
				locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
				rejectErr = h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			}

			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("[%s] can't delete storage place err: %w", handlerName, err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %w", handlerName, ticketInfo.TicketID, err)
	}
	return nil
}

func parseStoragePlacesFromDetail(detail string) []string {
	var storagePlaces []string
	parts := strings.TrimPrefix(detail, "МХ: ")
	spStrs := strings.Split(parts, ",")

	for _, spStr := range spStrs {
		spStr = strings.TrimSpace(spStr)
		if spStr != "" {
			storagePlaces = append(storagePlaces, spStr)
		}
	}

	return storagePlaces
}

type storagePlaceApiDeleter interface {
	DeleteStoragePlace(ctx context.Context, body deletestorageplacemodels.RequestDataForDeleteStoragePlace) error
}
