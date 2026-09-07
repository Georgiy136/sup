package creategoodssearchinventtask

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlers "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	creategoodssearchinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_goods_search_invent_task/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const handlerName = "TicketHandlerCreateGoodsSearchInventTask"

type TicketHandlerCreateGoodsSearchInventTask struct {
	repo                  services.HandlerTicketsRepo
	goodsOnStockTaskAdder goodsOnStockTaskAdder
}

func NewTicketHandlerCreateGoodsSearchInventTask(repo services.HandlerTicketsRepo, goodsOnStockTaskAdder goodsOnStockTaskAdder) *TicketHandlerCreateGoodsSearchInventTask {
	return &TicketHandlerCreateGoodsSearchInventTask{
		repo:                  repo,
		goodsOnStockTaskAdder: goodsOnStockTaskAdder,
	}
}

func (h *TicketHandlerCreateGoodsSearchInventTask) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext creategoodssearchinventtaskmodels.ExtTicketInfoCreateGoodsSearchInventTask

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
		TaskType:          commonhandlers.TaskTypeRLG,
	})
	if err != nil {
		return fmt.Errorf("[%s] can't create goods search invent task: %w", handlerName, err)
	}

	return nil
}

type goodsOnStockTaskAdder interface {
	AddGoodsOnStockTask(ctx context.Context, req commonhandlersmodels.HandlerRequestForAddGoodsOnStockTask) error
}
