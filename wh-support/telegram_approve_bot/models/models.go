package models

import ticketinfofmt "gitlab.wildberries.ru/wbwh/support/utils.git/ticket_info_formatter"

type Notification struct {
	TicketInfo       TicketInfo `json:"ticket_info" validate:"required"`
	Comments         string     `json:"comments" validate:"required"`
	ChatID           int64      `json:"tgchat_id" validate:"required"`
	TicketID         int64      `json:"ticket_id" validate:"required,gt=0"`
	EmployeeID       int64      `json:"employee_id" validate:"required,gt=0"`
	TypeOfEmployeeID string     `json:"type_of_employee_id" validate:"required,oneof=CRT GRP EMP FVR"`
}

type TicketInfo struct {
	Ext                   []ticketinfofmt.AddedInfo `json:"ext"`
	StatusId              string                    `json:"status_id"`
	TicketName            string                    `json:"ticket_name"`
	CategoryName          string                    `json:"category_name"`
	ParentCategoryName    string                    `json:"parent_category_name"`
	StatusDescription     string                    `json:"status_description"`
	ScenarioName          string                    `json:"scenario_name"`
	OperationType         string                    `json:"operation_type"`
	CreateEmployeeID      int64                     `json:"create_employee_id"`
	RejectedEmployeeID    *int64                    `json:"rejected_employee_id"`
	ResponsibleEmployeeID *int64                    `json:"responsible_employee_id"`
	NeedAddedInfo         *bool                     `json:"need_added_information_model"`
}

type TelegramExistInfo struct {
	IsExist bool `json:"is_exist"`
}

type EmployeeInfo struct {
	Name string `json:"employee_name"`
}

type CheckTelegramInfo struct {
	IsExist    bool  `json:"is_exist"`
	TgID       int64 `json:"tg_id"`
	EmployeeID int64 `json:"employee_id"`
}

type TicketData struct {
	Message          MessageInfo `json:"message"`
	Operation        string      `json:"operation"`
	TypeOfEmployeeID string      `json:"type_of_employee_id"`
}

type MessageInfo struct {
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
}
