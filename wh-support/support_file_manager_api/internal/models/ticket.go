package models

// TicketInfo - данные о заявке
type TicketInfo struct {
	TicketID         int64  `json:"ticket_id"`
	CategoryID       int64  `json:"category_id"`
	CreateEmployeeID int64  `json:"create_employee_id"`
	TicketType       string `json:"ticket_type"`
}
