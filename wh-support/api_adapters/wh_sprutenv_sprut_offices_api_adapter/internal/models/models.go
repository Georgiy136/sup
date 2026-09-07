package models

type Spruts struct {
	Data []Sprut `json:"data"`
}

type Sprut struct {
	Name string `json:"name"`
}

type RequestChesExistSprutOffice struct {
	OfficeID int64 `json:"office_id"`
}

type DataOfficeOnSprut struct {
	Data OfficeOnSprut `json:"data"`
}

type OfficeOnSprut struct {
	SprutName    string `json:"sprut_name"`
	OfficeID     int64  `json:"office_id"`
	OfficeExists bool   `json:"office_exists"`
}

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}

type ResponseGetOfficeWithInactive struct {
	Offices []struct {
		City           string  `json:"city"`
		FullAddress    string  `json:"full_address"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		OfficeId       int64   `json:"office_id"`
		OfficeName     string  `json:"office_name"`
		ParentOfficeId int64   `json:"parent_office_id"`
		LmRouteId      int64   `json:"lm_route_id"`
		TypePoint      int64   `json:"type_point"`
	} `json:"offices"`
}

type ResponseGetBranchOfficeInfo struct {
	Offices []struct {
		OfficeID            int64    `json:"office_id"`
		OfficeName          string   `json:"office_name"`
		CountryID           *int64   `json:"country_id,omitempty"`
		TypePoint           *int64   `json:"type_point,omitempty"`
		TypePointName       *string  `json:"type_point_name,omitempty"`
		Lat                 *float64 `json:"lat,omitempty"`
		Long                *float64 `json:"long,omitempty"`
		CountryCode         *string  `json:"country_code,omitempty"`
		FranchiseSupplierID *int64   `json:"franchise_supplier_id,omitempty"`
		PartnerSupplierID   *int64   `json:"partner_supplier_id,omitempty"`
		IsDeleted           bool     `json:"is_deleted"`
		IsPartner           bool     `json:"is_partner"`
		IsFranchise         bool     `json:"is_franchise"`
		IsExistsWarehouse   bool     `json:"is_exists_warehouse"`
		IsAsmExists         bool     `json:"is_asm_exists"`
		PartnerUserID       *string  `json:"partner_user_id,omitempty"`
		FranchiseUserID     *string  `json:"franchise_user_id,omitempty"`
		CloseDate           *string  `json:"close_date,omitempty"`
		FullAddress         *string  `json:"full_address,omitempty"`
		RmEmployeeID        *int64   `json:"rm_employee_id,omitempty"`
		DepRmEmployeeID     *int64   `json:"dep_rm_employee_id,omitempty"`
		Region              *string  `json:"region,omitempty"`
	} `json:"data"`
}

type ResponseGetBranchOffice struct {
	BranchOffices []BranchOffice `json:"value"`
	IsValid       bool           `json:"is_valid"`
	ErrMsg        string         `json:"err_msg"`
}

type BranchOffice struct {
	OfficeID      int64   `json:"office_id"`
	OfficeName    string  `json:"office_name"`
	TypePointName *string `json:"type_point_name"`
}
