package creategoodsinventtask

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlers "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	creategoodsinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_goods_invent_task/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const handlerName = "TicketHandlerCreateGoodsInventTask"

type TicketHandlerCreateGoodsInventTask struct {
	repo                  services.HandlerTicketsRepo
	goodsOnStockTaskAdder goodsOnStockTaskAdder
}

func NewTicketHandlerCreateGoodsInventTask(repo services.HandlerTicketsRepo, goodsOnStockTaskAdder goodsOnStockTaskAdder) *TicketHandlerCreateGoodsInventTask {
	return &TicketHandlerCreateGoodsInventTask{
		repo:                  repo,
		goodsOnStockTaskAdder: goodsOnStockTaskAdder,
	}
}

func (h *TicketHandlerCreateGoodsInventTask) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext creategoodsinventtaskmodels.ExtTicketInfoCreateGoodsInventTask

	err := handlerutils.DecodeMapToStructureWithoutErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}

	gosTasks := make([]commonhandlersmodels.GoodsOnStockTask, 0, len(ext.GoodsOnStockTasks))
	for i := range ext.GoodsOnStockTasks {
		gosTasks = append(gosTasks, commonhandlersmodels.GoodsOnStockTask{
			Barcode:   ext.GoodsOnStockTasks[i].Barcode.Value,
			GoodsID:   ext.GoodsOnStockTasks[i].GoodsID.Value,
			Priority:  ext.GoodsOnStockTasks[i].Priority.Value,
			ExtGoodID: ext.GoodsOnStockTasks[i].ExtGoodID.Value,
		})
	}

	err = h.goodsOnStockTaskAdder.AddGoodsOnStockTask(ctx, commonhandlersmodels.HandlerRequestForAddGoodsOnStockTask{
		OfficeID:          ext.OfficeID.ID,
		WhID:              ext.WhID.ID,
		EmployeeID:        ticketInfo.CreateEmployeeID,
		TicketID:          ticketInfo.TicketID,
		GoodsOnStockTasks: gosTasks,
		TaskType:          commonhandlers.TaskTypeIKZ,
	})
	if err != nil {
		return fmt.Errorf("[%s] can't create goods invent task: %w", handlerName, err)
	}

	return nil
}

type goodsOnStockTaskAdder interface {
	AddGoodsOnStockTask(ctx context.Context, req commonhandlersmodels.HandlerRequestForAddGoodsOnStockTask) error
}
