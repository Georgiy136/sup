package models

type BuildingsResponse struct {
	Data []BuildingInfo `json:"data"`
}

type BuildingInfo struct {
	OfficeID     int64  `json:"office_id"`
	ChEmployeeID int64  `json:"ch_employee_id"`
	BuildingID   string `json:"building_id"`
	ChDt         string `json:"ch_dt"`
	IsDefault    bool   `json:"is_default"`
}

type BuildingsSelector struct {
	BuildingID string `json:"id"`
}
