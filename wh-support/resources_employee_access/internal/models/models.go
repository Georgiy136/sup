package models

const (
	DatabaseKey          = "support_pgx"
	SystemResourcePrefix = "system"
)

type EmployeeResources struct {
	ChDt       string   `json:"ch_dt"`
	ActionIds  []string `json:"action_ids"`
	EmployeeId int64    `json:"employee_id"`
}

type EmployeeResourcesResp struct {
	ActionIds []string `json:"action_ids"`
}

type DBEmployeesByAction struct {
	ActionId     string               `json:"action_id"`
	EmployeeInfo []EmployeeInfoFromDB `json:"employee_info"`
}

type EmployeeInfoFromDB struct {
	ChDt         string `json:"ch_dt"`
	EmployeeId   int64  `json:"employee_id"`
	ChEmployeeID int64  `json:"ch_employee_id"`
}

type EmployeeInfo struct {
	EmployeeId int64  `json:"employee_id"`
	Name       string `json:"employee_name"`
}

type EmployeeInfoResp struct {
	ID             int64   `json:"employee_id"`
	Name           *string `json:"employee_name"`
	ChEmployeeID   int64   `json:"ch_employee_id"`
	ChEmployeeName *string `json:"ch_employee_name"`
	ChDt           string  `json:"ch_dt"`
}

type EmployeeAppAction struct {
	AppActionID   string `json:"appaction_id"`
	AppActionName string `json:"appaction_name"`
	IsMainView    bool   `json:"is_main_view"`
}
