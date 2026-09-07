package closewh

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/sirupsen/logrus"
)

const (
	deleteGoodsRetryIntervalMinutes = 60 * time.Minute
	deleteStoragePlaceTaskRedisTTL  = 24 * time.Hour

	deleteStoragePlaceTaskSentFlagKey = "wh_delete_storage_place_task:first_call:%d_%s"

	defaultRejectMsgUndeletedStoragePlace = "Остались неудаленные МХ"
)

func (t *TicketHandlerCloseWhNew) deleteStoragePlace(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForDeleteStoragePlace
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode ext: %w", err)
	}

	taskKey := fmt.Sprintf(deleteStoragePlaceTaskSentFlagKey, ticketInfo.TicketID, ticketInfo.StatusID)
	exist, err := t.redis.Exists(ctx, taskKey)
	if err != nil {
		return fmt.Errorf("redis check task key: %w", err)
	}
	action, resp, err := t.goodsProvider.DeleteGoodsFromStoragePlaces(ctx, closewhmodelsnew.RequestDeleteStoragePlace{
		WhId:       ext.WhId.Id,
		OfficeId:   ext.OfficeId.Id,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsRetry:    !exist,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("delete StoragePlace api err: %w", err)
	}
	if !exist {
		if setErr := t.redis.SetWithTTL(ctx, taskKey, "1", deleteStoragePlaceTaskRedisTTL); setErr != nil {
			logrus.Errorf("redis set task created: %v", setErr)
		}
	}

	logrus.Infof("ticket_id: %d, delete StoragePlace api resp: %+v", ticketInfo.TicketID, resp)

	switch action.Action {
	case closewhmodelsnew.ActionInProgress:
		nextCheckDt := time.Now().Add(deleteGoodsRetryIntervalMinutes)
		if err = t.repo.UpdateNextCheckAt(ctx, ticketInfo.TicketID, nextCheckDt); err != nil {
			return fmt.Errorf("can't update next check at for ticket %d: %w", ticketInfo.TicketID, err)
		}
		return nil
	case closewhmodelsnew.ActionCompleted:
		if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
			return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
		}
		if err = t.repo.PerformTicketWithScenario(ctx, ticketInfo.TicketID, nil, resolveStoragePlaceScenario(ticketInfo.StatusID, closewhmodelsnew.ActionCompleted, "")); err != nil {
			return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
		}
		return nil
	case closewhmodelsnew.ActionCompletedWithErrors:
		switch action.Reason {
		case closewhmodelsnew.WithErrReasonStoragePlacesNotDeleted:
			if isDeleteStoragePlaceRetryStatus(ticketInfo.StatusID) {
				rawErrValues, err := jsoniter.Marshal(models.PerformErrValue{Comment: defaultRejectMsgUndeletedStoragePlace})
				if err != nil {
					return fmt.Errorf("can't marshal performErr for ticket %d: %w", ticketInfo.TicketID, err)
				}
				locMsgs := localization.New(
					support_err_keys.KeyErrorStoragePlacesNotDeleted,
					map[string]string{"wh_id": strconv.FormatInt(ext.WhId.Id, 10)},
				)
				if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
					return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
				}
				nextScenario := resolveStoragePlaceScenario(ticketInfo.StatusID, closewhmodelsnew.ActionCompletedWithErrors, closewhmodelsnew.WithErrReasonStoragePlacesNotDeleted)
				if err = t.repo.PerformTicketWithScenarioAndErrComment(ctx, ticketInfo.TicketID, nil, nextScenario, rawErrValues, locMsgs); err != nil {
					return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
				}
				logrus.Infof("ticket %d performed with err comment (scenario %d) for %s/%s", ticketInfo.TicketID, nextScenario, statusIdDeleteStoragePlacePL1, statusIdDeleteStoragePlacePL2)
				return nil
			}

			nextScenarioOrderID := resolveUndeletedStoragePlacesScenario(resp.NotDeletedStoragePlace, ticketInfo.ReturnFromStatusID)
			var notDeletedStoragePlacesJSON []byte
			if nextScenarioOrderID == scenarioPlacesNotDeleted {
				notDeletedPlaces := buildNotDeletedPlacesPayload(resp.NotDeletedStoragePlace)
				notDeletedStoragePlacesJSON, err = jsoniter.Marshal(map[string]any{"places_ids": notDeletedPlaces})
				if err != nil {
					return fmt.Errorf("can't marshal not deleted places: %w", err)
				}
			}
			if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
				return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
			}
			if performErr := t.repo.PerformTicketWithScenario(ctx, ticketInfo.TicketID, notDeletedStoragePlacesJSON, nextScenarioOrderID); performErr != nil {
				return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, performErr)
			}
			return nil
		default:
			return fmt.Errorf("unknown completed with errors reason: %s", action.Reason)
		}
	case closewhmodelsnew.ActionFailed:
		return fmt.Errorf("delete storage place failed: %s", action.Reason)
	default:
		return fmt.Errorf("unknown delete storage place action: %s", action.Action)
	}
}

func (t *TicketHandlerCloseWhNew) getStoragePlaces(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForGetStoragePlaces
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode ext: %w", err)
	}

	bodyRequestForGetStoragePlaces := prepareRequestForGetStoragePlacesByWh(ext)

	placeIDs, err := t.whApi.GetStoragePlacesByWh(ctx, bodyRequestForGetStoragePlaces)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("get storage places by wh err: %w", err)
	}

	var extJSON []byte
	if len(placeIDs) != 0 {
		placeIDsPayload := buildPlaceIdsPayload(placeIDs)

		extJSON, err = jsoniter.Marshal(placeIDsPayload)
		if err != nil {
			return fmt.Errorf("can't marshal ext for ticket %d: %w", ticketInfo.TicketID, err)
		}
	}
	if err = t.repo.PerformTicket(ctx, ticketInfo.TicketID, extJSON); err != nil {
		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}
	return nil
}
