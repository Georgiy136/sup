package models

type NotificationRequest struct {
	TicketInfo       TicketInfo `json:"ticket_info"`
	Comments         string     `json:"comments"`
	TgChatID         int64      `json:"tgchat_id"`
	TicketID         int64      `json:"ticket_id"`
	EmployeeID       int64      `json:"employee_id"`
	TypeOfEmployeeID string     `json:"type_of_employee_id"`
}
