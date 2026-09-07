package commonhandlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	TaskTypeIKZ = "IKZ"
	TaskTypeKIZ = "KIZ"
	TaskTypeRLG = "RLG"

	invalidExciseExtIDComment = "Некорректный внешний идентификатор вещи (КИЗ) у одного или нескольких товаров. Список проблемных товаров сохранён в заявке."
)

type goodsOnStockAddErrText struct {
	comment string
	locKey  string
}

func errTextsByTaskType(taskType string) (goodsOnStockAddErrText, error) {
	switch taskType {
	case TaskTypeIKZ:
		return goodsOnStockAddErrText{
			comment: "Задания на инвентаризацию КИЗ созданы не для всех товаров. Список проблемных товаров сохранён в заявке.",
			locKey:  support_err_keys.KeyErrorInventTaskGoodsTaskNotCreated,
		}, nil
	case TaskTypeRLG:
		return goodsOnStockAddErrText{
			comment: "Задания на поиск товаров созданы не для всех товаров. Список проблемных товаров сохранён в заявке.",
			locKey:  support_err_keys.KeyErrorInventTaskGoodsSearchTaskNotCreated,
		}, nil
	default:
		return goodsOnStockAddErrText{}, fmt.Errorf("unsupported goods on stock task type %s", taskType)
	}
}

func (c *CommonHandlers) AddGoodsOnStockTask(ctx context.Context, req commonhandlersmodels.HandlerRequestForAddGoodsOnStockTask) error {
	errText, err := errTextsByTaskType(req.TaskType)
	if err != nil {
		return fmt.Errorf("can't get err text by tasktype: %w", err)
	}

	bodyRequest := prepareRequestForAddGoodsOnStockTask(req)

	goodsNotUpdated, err := c.inventApi.AddGoodsOnStockInventTask(ctx, bodyRequest)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on add goods on stock task: %v", req.TicketID, errWithMsg.Msg)
			return nil
		}
		if extIDs := parseInvalidExciseExtIDs(err.Error()); len(extIDs) > 0 {
			locMsgs := localization.New(support_err_keys.KeyErrorInventTaskInvalidExciseExtIDs, nil)
			rejectValues := map[string]any{
				"values": extIDs,
			}

			if rejectErr := c.repo.RejectTicketWithValues(ctx, req.TicketID, invalidExciseExtIDComment, locMsgs, rejectValues); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on invalid excise, ext_ids: %v", req.TicketID, extIDs)
			return nil
		}

		return fmt.Errorf("can't add goods on stock task: %w", err)
	}

	if len(goodsNotUpdated) > 0 {
		locMsgs := localization.New(errText.locKey, nil)

		if len(goodsNotUpdated) == len(req.GoodsOnStockTasks) {
			rejectValues := map[string]any{
				"values": goodsNotUpdated,
			}
			if rejectErr := c.repo.RejectTicketWithValues(ctx, req.TicketID, errText.comment, locMsgs, rejectValues); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on add goods on stock task: %s", req.TicketID, errText.comment)
			return nil
		}

		goodsNotUpdatedPayload := prepareGoodsNotUpdated(goodsNotUpdated)

		rawGoodsNotUpdated, err := jsoniter.Marshal(goodsNotUpdatedPayload)
		if err != nil {
			return fmt.Errorf("can't marshal goods not updated: %w", err)
		}

		rawErrComment, err := jsoniter.Marshal(models.PerformErrValue{Comment: errText.comment})
		if err != nil {
			return fmt.Errorf("can't marshal perform err comment: %w", err)
		}

		if err = c.repo.PerformTicketWithErrComment(ctx, req.TicketID, rawGoodsNotUpdated, rawErrComment, locMsgs); err != nil {
			locMsgs = localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("can't reject ticket %d on add goods on stock task: %v", req.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", req.TicketID, err)
		}

		logrus.Infof("ticket %d performed with err comment on add goods on stock task: %s", req.TicketID, errText.comment)
		return nil
	}

	if err = c.repo.PerformTicket(ctx, req.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d on add goods on stock task: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", req.TicketID, err)
	}

	return nil
}

func parseInvalidExciseExtIDs(errText string) []string {
	raw := handlerutils.ExtractValueFromText(errText, "ext_ids: [", "]")
	if raw == "" {
		return nil
	}

	return strings.Fields(raw)
}
