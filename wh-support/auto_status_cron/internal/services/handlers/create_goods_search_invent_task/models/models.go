package creategoodssearchinventtaskmodels

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

type ExtTicketInfoCreateGoodsSearchInventTask struct {
	OfficeID          OfficeID           `mapstructure:"office_id"`
	WhID              WhID               `mapstructure:"wh_id"`
	GoodsOnStockTasks []GoodsOnStockTask `mapstructure:"gos_tasks"`
}

type OfficeID = models.NamedID[int64]
type WhID = models.NamedID[int64]

type GoodsOnStockTask struct {
	Barcode   models.ValueField[string] `mapstructure:"barcode"`
	GoodsID   models.ValueField[int64]  `mapstructure:"goods_id"`
	Priority  models.ValueField[int64]  `mapstructure:"priority"`
	ExtGoodID models.ValueField[string] `mapstructure:"ext_good_id"`
}
