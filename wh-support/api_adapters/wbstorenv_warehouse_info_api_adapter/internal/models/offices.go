package models

type ListWarehousesOffices struct {
	Data []WarehousesOffices `json:"data"`
}

type WarehousesOffices struct {
	OfficeID          int64  `json:"office_id"`
	CountryCode       string `json:"country_code"`
	IsDeleted         bool   `json:"is_deleted"`
	HourOffsetFromMsk int64  `json:"hour_offset_from_msk"`
	TimeZone          string `json:"time_zone"`
	OfficeName        string `json:"office_name"`
	TypePoint         int64  `json:"type_point"`
}

type SelectorOffices struct {
	OfficeID   int64  `json:"id"`
	OfficeName string `json:"name"`
}

type ResponseGetSelectorBranchOffices struct {
	BranchOffices []BranchOffice `json:"value"`
	IsValid       bool           `json:"is_valid"`
	ErrMsg        string         `json:"err_msg"`
}

type BranchOffices struct {
	Data []BranchOffice `json:"data"`
}

type BranchOffice struct {
	OfficeId            int64  `json:"office_id"`
	OfficeName          string `json:"office_name"`
	IsDeleted           bool   `json:"is_deleted"`
	CountryId           int64  `json:"country_id"`
	IsPartner           bool   `json:"is_partner"`
	IsFranchisepvz      bool   `json:"is_franchisepvz"`
	FranchiseSupplierId int64  `json:"franchise_supplier_id"`
}
