package models

type GroupDataResponse struct {
	GroupID      int64          `json:"group_id"`
	EmployeeInfo []EmployeeData `json:"employee_info"`
}

type EmployeeData struct {
	ChDt         string `json:"ch_dt"`
	EmployeeID   int64  `json:"employee_id"`
	ChEmployeeID int64  `json:"ch_employee_id"`
}
