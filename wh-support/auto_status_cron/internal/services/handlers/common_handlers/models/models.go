package commonhandlersmodels

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

type HandlerRequestForCreateStages struct {
	OfficeID   int64
	WhID       int64
	Stages     []int64
	TicketID   int64
	EmployeeID int64
}

type HandlerRequestForCreateParts struct {
	OfficeID   int64
	WhID       int64
	Stages     []int64
	PartName   string
	TicketID   int64
	EmployeeID int64
}

type BodyRequestForCreateStage struct {
	OfficeID   int64 `json:"office_id"`
	WhId       int64 `json:"wh_id"`
	Stage      int64 `json:"stage"`
	EmployeeID int64 `json:"employee_id"`
}

type BodyRequestForCreatePart struct {
	EmployeeID int64  `json:"employee_id"`
	OfficeId   int64  `json:"office_id"`
	WhId       int64  `json:"wh_id"`
	Stage      int64  `json:"stage"`
	PartName   string `json:"part_name"`
}

type HandlerRequestForExclusionUpdateStreet struct {
	OfficeId               int64
	WhId                   int64
	IsExcludedFromAssembly bool
	Streets                []Streets
	EmployeeId             int64
	TicketID               int64
}

type BodyRequestForExclusionUpdateStreet struct {
	OfficeId               int64     `json:"office_id"`
	WhId                   int64     `json:"wh_id"`
	EmployeeId             int64     `json:"employee_id"`
	IsExcludedFromAssembly bool      `json:"is_excluded_from_assembly"`
	Streets                []Streets `json:"streets"`
}

type RequestForStageOrWhExclusionFromSaleUpdateV001 struct {
	WhId       int64   `json:"wh_id"`
	OfficeId   int64   `json:"office_id"`
	EmployeeId int64   `json:"employee_id"`
	IsExcluded bool    `json:"is_excluded"`
	Stages     []int64 `json:"stages"`
}

type HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001 struct {
	WhId       int64
	OfficeId   int64
	EmployeeId int64
	IsExcluded bool
	Stages     []int64
	TicketID   int64
}

type RequestForStageOrWhExclusionFromAssemblyUpdateV001 struct {
	WhId       int64   `json:"wh_id"`
	OfficeId   int64   `json:"office_id"`
	EmployeeId int64   `json:"employee_id"`
	IsExcluded bool    `json:"is_excluded"`
	Stages     []int64 `json:"stages"`
}

type Streets struct {
	Stage  int64 `json:"stage"`
	Street int64 `json:"street"`
}

type HandlerRequestForStreetExclusionFromSaleUpdateV002 struct {
	OfficeId           int64
	WhId               int64
	IsExcludedFromSale bool
	Streets            []StreetsWithSections
	EmployeeId         int64
	TicketID           int64
	ReplaceOrders      bool
}

type BodyRequestForStreetExclusionFromSaleUpdate struct {
	OfficeId               int64                 `json:"office_id"`
	WhId                   int64                 `json:"wh_id"`
	EmployeeId             int64                 `json:"employee_id"`
	IsExcludedFromAssembly bool                  `json:"is_excluded"`
	Streets                []StreetsWithSections `json:"streets" validate:"dive"`
}

type BodyRequestForStreetExclusionFromSaleUpdateV002 struct {
	OfficeId               int64                 `json:"office_id"`
	WhId                   int64                 `json:"wh_id"`
	EmployeeId             int64                 `json:"employee_id"`
	IsExcludedFromAssembly bool                  `json:"is_excluded"`
	Streets                []StreetsWithSections `json:"streets" validate:"dive"`
	ReplaceOrders          bool                  `json:"replace_orders"`
}

type StreetsWithSections struct {
	Stage       int64 `json:"stage"`
	Street      int64 `json:"street"`
	SectionFrom int64 `json:"section_from"`
	SectionTo   int64 `json:"section_to" validate:"gtefield=SectionFrom"`
}

type HandlerRequestForAddGoodsOnStockTask struct {
	OfficeID          int64
	WhID              int64
	EmployeeID        int64
	TicketID          int64
	GoodsOnStockTasks []GoodsOnStockTask
	TaskType          string
}

type GoodsOnStockTask struct {
	Barcode   string
	GoodsID   int64
	Priority  int64
	ExtGoodID string
}

type BodyRequestForAddGoodsOnStockTask struct {
	OfficeID int64            `json:"office_id"`
	WhID     int64            `json:"wh_id"`
	GosTasks []RequestGosTask `json:"gos_tasks"`
}

type RequestGosTask struct {
	GoodsID    int64          `json:"goods_id"`
	EmployeeID int64          `json:"employee_id"`
	Barcode    string         `json:"barcode,omitempty"`
	ExtIDs     []RequestExtID `json:"ext_ids,omitempty"`
	Dt         string         `json:"dt"`
	TaskType   string         `json:"task_type"`
	Priority   int64          `json:"priority"`
}

type RequestExtID struct {
	ExtID     string `json:"ext_id"`
	ExtTypeID string `json:"ext_type_id"`
	IsCorrect bool   `json:"is_correct"`
}

type GoodsNotUpdatedPayload struct {
	GoodsNotUpdated []GoodsNotUpdatedItem `json:"bad_goods_ids"`
}

type GoodsNotUpdatedItem struct {
	GoodsID models.ValueField[int64] `json:"goods_id"`
}
