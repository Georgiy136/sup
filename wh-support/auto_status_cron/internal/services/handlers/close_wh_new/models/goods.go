package closewhmodels

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

type ExtTicketInfoForGoodsTask struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type GoodsItem struct {
	PlaceId int64  `json:"place_id"`
	GoodsId int64  `json:"goods_id"`
	SkuId   *int64 `json:"sku_id"`
	NmId    *int64 `json:"nm_id"`
}

type GoodsItemWithPrice struct {
	GoodsId    int64   `json:"goods_id"`
	PlaceId    int64   `json:"place_id"`
	SkuId      *int64  `json:"sku_id"`
	NmId       *int64  `json:"nm_id"`
	AvgNmPrice float64 `json:"avg_nm_price"`
}

type GoodsTaskData struct {
	WhId   int64           `json:"wh_id"`
	Status GoodsTaskStatus `json:"status"`
}

type GoodsTaskStatus string

const (
	GoodsTaskStatusPND GoodsTaskStatus = "PND" // В обработке
	GoodsTaskStatusPRG GoodsTaskStatus = "PRG" // В процессе
	GoodsTaskStatusERR GoodsTaskStatus = "ERR" // Ошибка
	GoodsTaskStatusCMP GoodsTaskStatus = "CMP" // Завершено
)

type ResponseGoodsTask struct {
	Data GoodsTaskData `json:"data"`
}

type RequestCreateGoodsTask struct {
	WhId       int64 `json:"wh_id"`
	OfficeId   int64 `json:"office_id"`
	EmployeeId int64 `json:"employee_id"`
	IsRetry    bool  `json:"is_retry,omitempty"`
}

type RequestGetGoodsPage struct {
	WhId        int64 `json:"wh_id"`
	OfficeId    int64 `json:"office_id"`
	LastGoodsId int64 `json:"last_goods_id"`
}

type ResponseGoodsPage struct {
	Data []GoodsItem `json:"data"`
}

type RequestPriceAggregation struct {
	NmIds []int64 `json:"list"`
}

type PriceItem struct {
	NmId     int64   `json:"nm_id"`
	AvgPrice float64 `json:"avg_price"`
}

type ResponsePriceAggregation struct {
	Data []PriceItem `json:"data"`
}

type RequestDeleteGoodsData struct {
	WhId        int64 `json:"wh_id"`
	OfficeId    int64 `json:"office_id"`
	LastGoodsId int64 `json:"last_goods_id"`
}

type TotalCountGoodsDBSaveData struct {
	TotalCountGoodsID int64 `json:"total_count_goods_id"`
}

type GoodsWithPriceDBSaveData struct {
	GoodsID           []models.Files `json:"goods_id"`
	TotalPriceSum     float64        `json:"total_price_sum"`
	TotalCountGoodsID int64          `json:"total_count_goods_id"`
}
