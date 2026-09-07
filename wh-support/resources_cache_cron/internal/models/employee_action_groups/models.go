package employeeactiongroups

import "time"

type EmployeeActionGroupEvent struct {
	EmployeeId int64     `json:"employee_id"`
	GroupId    int64     `json:"group_id"`
	IsDel      bool      `json:"is_del"`
	ChDt       string    `json:"ch_dt"`
	ChDtParsed time.Time `json:"-"`
}

type EmployeeActionGroupDBResp struct {
	LogId        int64                      `json:"log_id"`
	SleepSeconds int64                      `json:"sleep_seconds"`
	Data         []EmployeeActionGroupEvent `json:"data"`
}
