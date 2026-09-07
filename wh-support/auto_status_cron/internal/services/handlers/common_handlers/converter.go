package commonhandlers

import (
	"time"

	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

func prepareRequestForCreateStage(whId, officeId, stage int64) commonhandlersmodels.BodyRequestForCreateStage {
	return commonhandlersmodels.BodyRequestForCreateStage{
		WhId:     whId,
		OfficeID: officeId,
		Stage:    stage,
	}
}

func prepareRequestForCreateParts(officeId, whId, stage int64, partName string) commonhandlersmodels.BodyRequestForCreatePart {
	return commonhandlersmodels.BodyRequestForCreatePart{
		OfficeId: officeId,
		WhId:     whId,
		Stage:    stage,
		PartName: partName,
	}
}

func prepareRequestDataForExcludeStreetForAssembly(in commonhandlersmodels.HandlerRequestForExclusionUpdateStreet) commonhandlersmodels.BodyRequestForExclusionUpdateStreet {
	return commonhandlersmodels.BodyRequestForExclusionUpdateStreet{
		OfficeId:               in.OfficeId,
		WhId:                   in.WhId,
		IsExcludedFromAssembly: in.IsExcludedFromAssembly,
		Streets:                in.Streets,
		EmployeeId:             in.EmployeeId,
	}
}

func prepareRequestDataForStreetExclusionFromSaleUpdateV002(in commonhandlersmodels.HandlerRequestForStreetExclusionFromSaleUpdateV002) commonhandlersmodels.BodyRequestForStreetExclusionFromSaleUpdateV002 {
	return commonhandlersmodels.BodyRequestForStreetExclusionFromSaleUpdateV002{
		OfficeId:               in.OfficeId,
		WhId:                   in.WhId,
		IsExcludedFromAssembly: in.IsExcludedFromSale,
		Streets:                in.Streets,
		EmployeeId:             in.EmployeeId,
		ReplaceOrders:          in.ReplaceOrders,
	}
}

func prepareRequestDataForStageOrWhExclusionFromAssemblyUpdateV001(in commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001) commonhandlersmodels.RequestForStageOrWhExclusionFromAssemblyUpdateV001 {
	return commonhandlersmodels.RequestForStageOrWhExclusionFromAssemblyUpdateV001{
		WhId:       in.WhId,
		OfficeId:   in.OfficeId,
		EmployeeId: in.EmployeeId,
		IsExcluded: in.IsExcluded,
		Stages:     in.Stages,
	}
}

func prepareRequestForAddGoodsOnStockTask(req commonhandlersmodels.HandlerRequestForAddGoodsOnStockTask) commonhandlersmodels.BodyRequestForAddGoodsOnStockTask {
	gosTasks := make([]commonhandlersmodels.RequestGosTask, 0, len(req.GoodsOnStockTasks))

	for i := range req.GoodsOnStockTasks {
		task := commonhandlersmodels.RequestGosTask{
			GoodsID:    req.GoodsOnStockTasks[i].GoodsID,
			EmployeeID: req.EmployeeID,
			Barcode:    req.GoodsOnStockTasks[i].Barcode,
			Dt:         time.Now().Format(time.RFC3339Nano),
			TaskType:   req.TaskType,
			Priority:   req.GoodsOnStockTasks[i].Priority,
		}

		if req.GoodsOnStockTasks[i].ExtGoodID != "" {
			task.ExtIDs = []commonhandlersmodels.RequestExtID{
				{
					ExtID:     req.GoodsOnStockTasks[i].ExtGoodID,
					ExtTypeID: TaskTypeKIZ,
					IsCorrect: true,
				},
			}
		}

		gosTasks = append(gosTasks, task)
	}

	return commonhandlersmodels.BodyRequestForAddGoodsOnStockTask{
		OfficeID: req.OfficeID,
		WhID:     req.WhID,
		GosTasks: gosTasks,
	}
}

func prepareGoodsNotUpdated(goodsIDs []int64) commonhandlersmodels.GoodsNotUpdatedPayload {
	const (
		goodsIDOrderID       = 1
		goodsIDFrontDataName = "goods_id"
	)

	items := make([]commonhandlersmodels.GoodsNotUpdatedItem, 0, len(goodsIDs))
	for i := range goodsIDs {
		items = append(items, commonhandlersmodels.GoodsNotUpdatedItem{
			GoodsID: models.ValueField[int64]{
				Value:         goodsIDs[i],
				OrderID:       goodsIDOrderID,
				FrontDataName: goodsIDFrontDataName,
			},
		})
	}

	return commonhandlersmodels.GoodsNotUpdatedPayload{GoodsNotUpdated: items}
}
