package models

type EmployeeResources struct {
	ChDt       string   `json:"ch_dt"`
	ActionIds  []string `json:"action_ids"`
	EmployeeId int64    `json:"employee_id"`
}

type ListEmployeeResources struct {
	AppActions []AppActions `json:"app_actions"`
}

type AppActions struct {
	Action string `json:"appaction_name"`
}
