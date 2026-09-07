package models

const (
	GetNotificationsSP = "sync.chatnotification_exporttojson"
	DatabaseKey        = "support_pgx"
)

type DBTicketsForNotificationResponse struct {
	Data         []DBTicketForNotification `json:"data"`
	LogID        int64                     `json:"log_id"`
	SleepSeconds int64                     `json:"sleep_seconds"`
}

type DBTicketForNotification struct {
	TicketInfo       TicketInfo `json:"ticket_info"`
	Comments         string     `json:"comments"`
	TgChatID         int64      `json:"tgchat_id"`
	TicketID         int64      `json:"ticket_id"`
	EmployeeID       int64      `json:"employee_id"`
	TypeOfEmployeeID string     `json:"type_of_employee_id"`
	Dt               string     `json:"ch_dt"`
}

type TicketInfo struct {
	Ext                   []AddedInfo `json:"ext"`
	StatusId              string      `json:"status_id"`
	TicketName            string      `json:"ticket_name"`
	CategoryName          string      `json:"category_name"`
	ParentCategoryName    string      `json:"parent_category_name"`
	StatusDescription     string      `json:"status_description"`
	ScenarioName          string      `json:"scenario_name"`
	OperationType         string      `json:"operation_type"`
	CreateEmployeeID      int64       `json:"create_employee_id"`
	RejectedEmployeeID    *int64      `json:"rejected_employee_id"`
	ResponsibleEmployeeID *int64      `json:"responsible_employee_id"`
	NeedAddedInfo         *bool       `json:"need_added_information_model"`
}

type AddedInfo struct {
	Value         interface{} `json:"value"`
	OrderID       int64       `json:"order_id"`
	FrontDataName string      `json:"front_data_name"`
}
