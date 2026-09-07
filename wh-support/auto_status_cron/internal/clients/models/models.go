package models

type EmployeeInfo struct {
	Name string `json:"employee_name"`
}

type DataWrapper[T any] struct {
	Data T `json:"data"`
}

type AddGoodsOnStockTaskDataResponse struct {
	GoodsNotUpdated []int64 `json:"goods_not_updated"`
}

type CreateStoragePlaceInventTaskData struct {
	CreatedQty int64 `json:"created_qty"`
}
