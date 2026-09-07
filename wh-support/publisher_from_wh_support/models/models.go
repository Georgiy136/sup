package models

type TicketsFromDB struct {
	Data []struct {
		Ext               map[string]interface{} `json:"ext"`
		ChDt              string                 `json:"ch_dt"`
		ChEmployeeID      int64                  `json:"ch_employee_id"`
		CreateDt          string                 `json:"create_dt"`
		TicketID          int64                  `json:"ticket_id"`
		CategoryID        int64                  `json:"category_id"`
		StatusId          string                 `json:"status_id"`
		Comments          *string                `json:"comments"`
		RejectedComments  *string                `json:"rejected_comments"`
		CreateEmployeeID  int64                  `json:"create_employee_id"`
		ApproveEmployeeID *int64                 `json:"approve_employee_id"`
		PerformEmployeeID *int64                 `json:"perform_employee_id"`
	} `json:"data"`
	LogID        int64 `json:"log_id"`
	SleepSeconds int64 `json:"sleep_seconds"`
}
