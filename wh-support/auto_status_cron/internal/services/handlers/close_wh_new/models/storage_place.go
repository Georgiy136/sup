package closewhmodels

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

type Reason string

type NotDeletedStoragePlace struct {
	Reason    Reason  `json:"reason"`
	PlaceIds  []int64 `json:"place_ids"`
	IsStock   bool    `json:"is_stock"`
	PlacesQty int64   `json:"places_qty"`
}

const (
	ReasonStockRemains       Reason = "STK"
	ReasonTareWithContent    Reason = "TAR"
	ReasonInventoryTask      Reason = "INV"
	ReasonUnfinishedTask     Reason = "TSK"
	ReasonWarehouseTransport Reason = "TRN"
	ReasonSortingSquare      Reason = "SRT"
)

var reasonDescriptionMap = map[Reason]string{
	ReasonStockRemains:       "На МХ есть остатки товаров",
	ReasonTareWithContent:    "На неосновном МХ припаркована тара с содержимым",
	ReasonInventoryTask:      "На основном МХ есть задание на инвентаризацию",
	ReasonUnfinishedTask:     "На МХ есть незавершённая задача",
	ReasonWarehouseTransport: "На МХ припаркован складской транспорт (тип 1623)",
	ReasonSortingSquare:      "К МХ привязан квадрат сортировки (тип 1505)",
}

func ReasonToString(r Reason) string {
	if desc, ok := reasonDescriptionMap[r]; ok {
		return desc
	}
	return string(r)
}

type DeleteStoragePlaceData struct {
	WhId                   int64                    `json:"wh_id"`
	Status                 DeleteStoragePlaceStatus `json:"status"`
	NotDeletedStoragePlace []NotDeletedStoragePlace `json:"notdeleted_places"`
}

type DeleteStoragePlaceStatus string

const (
	DeleteStoragePlaceStatusPND DeleteStoragePlaceStatus = "PND" // В обработке
	DeleteStoragePlaceStatusPRG DeleteStoragePlaceStatus = "PRG" // В процессе
	DeleteStoragePlaceStatusDEL DeleteStoragePlaceStatus = "DEL" // Удалено
	DeleteStoragePlaceStatusERR DeleteStoragePlaceStatus = "ERR" // Ошибка
)

type RequestDeleteStoragePlace struct {
	WhId       int64 `json:"wh_id"`
	OfficeId   int64 `json:"office_id"`
	EmployeeId int64 `json:"employee_id"`
	IsRetry    bool  `json:"is_retry,omitempty"`
}

type ExtTicketInfoForDeleteStoragePlace struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type ExtTicketInfoForGetStoragePlaces struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type StoragePlaceIdsDBSaveData struct {
	Reason   models.FieldValue `json:"reason"`
	Count    models.FieldValue `json:"count"`
	PlaceIds models.FieldValue `json:"place_ids"`
}

type PlaceIdsPayload struct {
	PlaceIds []PlaceIdItem `json:"place_ids_for_approve"`
}

type PlaceIdItem struct {
	PlaceId models.FieldValue `json:"place_id"`
}
