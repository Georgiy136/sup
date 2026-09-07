package models

type ChatEventNotification struct {
	Event              string  `json:"event"`
	TicketID           int64   `json:"ticket_id"`
	Message            string  `json:"message"`
	MessageID          string  `json:"message_id"`
	ChatID             string  `json:"chat_id"`
	SenderEmployeeID   int64   `json:"sender_employee_id"`
	SenderEmployeeName string  `json:"sender_employee_name"`
	NotifyEmployeeIDs  []int64 `json:"notify_employee_ids"`
	TaggedEmployeeIDs  []int64 `json:"tagged_employee_ids"`
	CreatedDt          string  `json:"created_dt"`
}
