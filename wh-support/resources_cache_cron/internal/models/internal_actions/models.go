package internalactions

type EmployeeResources struct {
	ChDt         string   `json:"ch_dt"`
	ActionIds    []string `json:"action_ids"`
	EmployeeId   int64    `json:"employee_id"`
	ChEmployeeID int64    `json:"ch_employee_id"`
}

type EmployeeResourcesDBResp struct {
	LogId        int64                        `json:"log_id"`
	SleepSeconds int64                        `json:"sleep_seconds"`
	Data         []EmployeeResourcesWithLogID `json:"data"`
}

type EmployeeResourcesWithLogID struct {
	ChDt         string `json:"ch_dt"`
	ActionId     string `json:"action_id"`
	EmployeeId   int64  `json:"employee_id"`
	ChEmployeeID int64  `json:"ch_employee_id"`
	IsDel        bool   `json:"is_del"`
}
