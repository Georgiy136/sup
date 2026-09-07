package models

const DatabaseKey = "support_pgx"

type EmployeeInfo struct {
	ID   int64  `json:"employee_id"`
	Name string `json:"employee_name"`
}

type DBEmployeesByGroup struct {
	GroupID      int64 `json:"group_id"`
	EmployeeInfo []struct {
		ChDt         string `json:"ch_dt"`
		EmployeeID   int64  `json:"employee_id"`
		ChEmployeeID int64  `json:"ch_employee_id"`
	} `json:"employee_info"`
}

type EmployeeInfoResp struct {
	ChDt           string  `json:"ch_dt"`
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   *string `json:"employee_name"`
	ChEmployeeID   int64   `json:"ch_employee_id"`
	ChEmployeeName *string `json:"ch_employee_name"`
}
