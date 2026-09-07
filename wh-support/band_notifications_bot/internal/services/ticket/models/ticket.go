package models

import ticketinfofmt "gitlab.wildberries.ru/wbwh/support/utils.git/ticket_info_formatter"

type TicketNotifications struct {
	Data         []TicketNotification `json:"data"`
	LogID        int64                `json:"log_id"`
	SleepSeconds int64                `json:"sleep_seconds"`
}

type TicketNotification struct {
	TicketInfo       TicketInfo     `json:"ticket_info"`
	Comments         string         `json:"comments"`
	TicketID         int64          `json:"ticket_id"`
	EmployeeID       int64          `json:"employee_id"`
	TypeOfEmployeeID TypeOfEmployee `json:"type_of_employee_id"`
	Dt               string         `json:"ch_dt"`
}

type TicketInfo struct {
	Ext                       []ticketinfofmt.AddedInfo `json:"ext"`
	StatusID                  string                    `json:"status_id"`
	CategoryID                int64                     `json:"category_id"`
	TicketName                string                    `json:"ticket_name"`
	CategoryName              string                    `json:"category_name"`
	ScenarioName              string                    `json:"scenario_name"`
	OperationType             string                    `json:"operation_type"`
	CreateEmployeeID          int64                     `json:"create_employee_id"`
	StatusDescription         string                    `json:"status_description"`
	ParentCategoryName        string                    `json:"parent_category_name"`
	RejectedEmployeeID        *int64                    `json:"rejected_employee_id"`
	ResponsibleEmployeeID     *int64                    `json:"responsible_employee_id"`
	NeedAddedInformationModel bool                      `json:"need_added_information_model"`
	IsBlockedReject           bool                      `json:"is_blocked_reject"`
	IsBlockedStatus           bool                      `json:"is_blocked_status"`
	ReturnStatuses            []ReturnStatus            `json:"return_statuses"`
}

type ReturnStatus struct {
	ReturnStatusID          string `json:"return_status_id"`
	ReturnStatusDescription string `json:"return_status_description"`
}

type TicketPost struct {
	TicketID         int64          `json:"ticket_id"`
	CategoryID       int64          `json:"category_id"`
	PostID           string         `json:"post_id"`
	ChannelID        string         `json:"channel_id"`
	Operation        string         `json:"operation"`
	TypeOfEmployeeID TypeOfEmployee `json:"type_of_employee_id"`
	EmployeeID       int64          `json:"employee_id"`
}

type ActionOptions struct {
	BlockReject       bool
	BlockStatus       bool
	HasReturnStatuses bool
}
