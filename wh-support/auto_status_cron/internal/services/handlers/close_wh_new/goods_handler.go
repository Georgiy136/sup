package closewh

import (
	"context"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/consts"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/sirupsen/logrus"
)

const (
	goodsTaskRetryNextCheckMinutes = 30 * time.Minute
	goodsTaskRedisTTL              = 24 * time.Hour

	goodsTaskSentFlagKey      = "wh_goods_task:first_call:%d_%s"
	goodsExportFileNameFormat = "wh_%d.xlsx"

	errGoodsExceedsMaxTotal = "Количество остатков более %d"
)

func (t *TicketHandlerCloseWhNew) getGoods(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForGoodsTask
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode ext: %w", err)
	}

	taskKey := fmt.Sprintf(goodsTaskSentFlagKey, ticketInfo.TicketID, ticketInfo.StatusID)
	exist, err := t.redis.Exists(ctx, taskKey)
	if err != nil {
		return fmt.Errorf("redis check task key: %w", err)
	}

	action, err := t.goodsProvider.GoodsExportTaskCreate(ctx, closewhmodelsnew.RequestCreateGoodsTask{
		WhId:       ext.WhId.Id,
		OfficeId:   ext.OfficeId.Id,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsRetry:    !exist,
	})
	if err != nil {
		return fmt.Errorf("create goods task err: %w", err)
	}

	if !exist {
		if setErr := t.redis.SetWithTTL(ctx, taskKey, "1", goodsTaskRedisTTL); setErr != nil {
			return fmt.Errorf("redis set task created: %w", setErr)
		}
	}
	logrus.Infof("ticket_id: %d,get goods api resp: %+v", ticketInfo.TicketID, action)

	switch action.Action {
	case closewhmodelsnew.ActionInProgress:
		nextCheckDt := time.Now().Add(goodsTaskRetryNextCheckMinutes)
		if updateErr := t.repo.UpdateNextCheckAt(ctx, ticketInfo.TicketID, nextCheckDt); updateErr != nil {
			return fmt.Errorf("can't update next check at for ticket %d: %w", ticketInfo.TicketID, updateErr)
		}
		logrus.Infof("ticket %d goods task retry, next check at %s", ticketInfo.TicketID, nextCheckDt.Format(time.RFC3339Nano))
		return nil
	case closewhmodelsnew.ActionFailed:
		if action.Reason == closewhmodelsnew.FailReasonLimitExceeded {
			rawErrValues, err := jsoniter.Marshal(models.PerformErrValue{Comment: fmt.Sprintf(errGoodsExceedsMaxTotal, goodsMaxTotal)})
			if err != nil {
				return fmt.Errorf("can't marshal performErr for ticket %d: %w", ticketInfo.TicketID, err)
			}
			locMsgs := localization.New(
				support_err_keys.KeyWrnGoodsLimitExceeded,
				map[string]string{"limit": strconv.FormatInt(goodsMaxTotal, 10)},
			)
			if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
				return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
			}
			if err = t.repo.PerformTicketWithScenarioAndErrComment(ctx, ticketInfo.TicketID, nil, resolveGoodsScenario(closewhmodelsnew.ActionFailed, closewhmodelsnew.FailReasonLimitExceeded, 0, 0), rawErrValues, locMsgs); err != nil {
				return fmt.Errorf("can't perform ticket %d with err: %w", ticketInfo.TicketID, err)
			}
			return nil
		}
		return fmt.Errorf("goods task failed: %s", action.Reason)
	case closewhmodelsnew.ActionCompleted:
	}

	allGoods, lastGoodsID, err := t.goodsProvider.FetchAllGoods(ctx, ext.WhId.Id, ext.OfficeId.Id)
	if err != nil {
		return fmt.Errorf("can't fetch all goods: %w", err)
	}

	if len(allGoods) > goodsMaxTotal {
		rawErrValues, err := jsoniter.Marshal(closewhmodelsnew.TotalCountGoodsDBSaveData{TotalCountGoodsID: int64(len(allGoods))})
		if err != nil {
			return fmt.Errorf("can't marshal performErr for ticket %d: %w", ticketInfo.TicketID, err)
		}
		if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
			return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
		}
		if err = t.repo.PerformTicketWithScenario(ctx, ticketInfo.TicketID, rawErrValues, resolveGoodsScenario(closewhmodelsnew.ActionCompleted, "", int64(len(allGoods)), 0)); err != nil {
			return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
		}
		if err = t.goodsProvider.DeleteGoodsByWh(ctx, closewhmodelsnew.RequestDeleteGoodsData{
			WhId:        ext.WhId.Id,
			OfficeId:    ext.OfficeId.Id,
			LastGoodsId: lastGoodsID,
		}); err != nil {
			return fmt.Errorf("can't delete goods data: %w", err)
		}
		return nil
	}

	annotatedGoods, totalPriceSum, err := t.goodsProvider.AnnotateGoodsWithPrices(ctx, allGoods)
	if err != nil {
		return fmt.Errorf("can't annotate goods with prices: %w", err)
	}

	sheetData := buildGoodsWithPriceSheetData(annotatedGoods)
	excelBytes, err := t.docGen.GenerateExcel(sheetData)
	if err != nil {
		return fmt.Errorf("can't generate goods excel: %w", err)
	}

	fileName := fmt.Sprintf(goodsExportFileNameFormat, ext.WhId.Id)
	fileID, err := t.fileManagerApi.UploadFile(ctx, excelBytes, fileName)
	if err != nil {
		return fmt.Errorf("can't upload excel: %w", err)
	}

	fileExcel := []models.Files{{FileID: fileID, FileType: excelMimeType, FileSize: int64(len(excelBytes)), FileName: fileName}}
	filesJSON, err := jsoniter.Marshal(map[string]any{"goods_id": fileExcel})
	if err != nil {
		return fmt.Errorf("can't marshal file raw values: %w", err)
	}

	extJSON, err := jsoniter.Marshal(closewhmodelsnew.GoodsWithPriceDBSaveData{
		GoodsID:           fileExcel,
		TotalPriceSum:     totalPriceSum,
		TotalCountGoodsID: int64(len(allGoods)),
	})
	if err != nil {
		return fmt.Errorf("can't marshal ext raw values: %w", err)
	}
	if delErr := t.redis.Del(ctx, taskKey); delErr != nil {
		return fmt.Errorf("delete redis task key: action=%s reason=%s: %w", action.Action, action.Reason, delErr)
	}
	if err = t.ticketApi.PerformTicketV2(ctx, models.RequestPerformTicketV2{
		TicketID:        ticketInfo.TicketID,
		ScenarioOrderID: resolveGoodsScenario(closewhmodelsnew.ActionCompleted, "", int64(len(allGoods)), totalPriceSum),
		Ext:             extJSON,
		Files:           filesJSON,
	}, consts.SystemEmployeeID); err != nil {
		return fmt.Errorf("can't perform ticket with files: %w", err)
	}

	if err = t.goodsProvider.DeleteGoodsByWh(ctx, closewhmodelsnew.RequestDeleteGoodsData{
		WhId:        ext.WhId.Id,
		OfficeId:    ext.OfficeId.Id,
		LastGoodsId: lastGoodsID,
	}); err != nil {
		return fmt.Errorf("can't delete goods data: %w", err)
	}
	return nil
}
