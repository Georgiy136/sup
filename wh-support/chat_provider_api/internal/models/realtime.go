package models

type RealtimePostEvent struct {
	Message      string
	MessageID    string
	ChatID       string
	TicketID     int64
	EmployeeID   int64
	EmployeeName string
	CreatedDt    string
	Files        []FileInfo
}
