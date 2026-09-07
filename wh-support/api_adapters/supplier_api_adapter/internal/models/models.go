package models

type SupplierInfo struct {
	NmID         int64  `json:"nm_id"`
	SupplierID   int64  `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`
}

type GetSupplierInfoByNmIDsResponse struct {
	SupplierInfo []SupplierInfo `json:"value"`
	IsValid      bool           `json:"is_valid"`
	ErrMsg       string         `json:"err_msg"`
}
