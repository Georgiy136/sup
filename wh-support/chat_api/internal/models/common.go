package models

type BandChat struct {
	TicketID   int64  `json:"ticket_id"`
	ChatID     string `json:"chat_id"`
	EmployeeID int64  `json:"employee_id"`
}

type Chat struct {
	ChatID        string                 `json:"id"`
	ChatName      string                 `json:"chat_name"`
	EmployeeID    int64                  `json:"employee_id"`
	CreatedDt     string                 `json:"created_dt"`
	Created       bool                   `json:"created"`
	RootMessage   ChatMessage            `json:"root_message"`
	MessagesOrder []string               `json:"messages_order"`
	Messages      map[string]ChatMessage `json:"messages"`
}

type TicketInfo struct {
	ChatID               *string `json:"chat_id"`
	TicketID             int64   `json:"ticket_id"`
	CategoryID           int64   `json:"category_id"`
	CreateEmployeeID     int64   `json:"create_employee_id"`
	GroupEmployeeIDs     []int64 `json:"group_employee_ids"`
	FavouriteEmployeeIDs []int64 `json:"favourite_employee_ids"`
}

type ChatMessage struct {
	ID           string `json:"id"`
	Message      string `json:"message"`
	EmployeeID   int64  `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	Files        []File `json:"files"`
	CreatedDt    string `json:"created_dt"`
}

type File struct {
	FileID    string `json:"file_id"`
	FileName  string `json:"file_name"`
	MessageID string `json:"message_id"`
	MimeType  string `json:"mime_type"`
}

type TicketEmployeeIDsForTag struct {
	ChatID      string  `json:"chat_id"`
	TicketID    int64   `json:"ticket_id"`
	EmployeeIDs []int64 `json:"employee_ids"`
}
