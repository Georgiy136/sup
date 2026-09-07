package models

type WarehouseInfo struct {
	WhID             int64  `json:"wh_id"`
	WhName           string `json:"wh_name"`
	IsNotActive      bool   `json:"is_not_active"`
	IsExistVirtualWh bool   `json:"is_exist_virtual_wh"`
}

type GetWarehouseInfoByWhIDResponse struct {
	WarehouseInfo []WarehouseInfo `json:"value"`
	IsValid       bool            `json:"is_valid"`
	ErrMsg        string          `json:"err_msg"`
}
