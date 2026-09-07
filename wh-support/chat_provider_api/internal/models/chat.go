package models

type CreateChatResponse struct {
	TicketID int64    `json:"ticket_id"`
	Chat     ChatMeta `json:"chat"`
}

type ChatMeta struct {
	ID         string `json:"id"`
	ChatName   string `json:"chat_name"`
	EmployeeID int64  `json:"employee_id"`
	CreatedDt  string `json:"created_dt"`
}

type FileInfo struct {
	FileID    string `json:"file_id"`
	FileName  string `json:"file_name"`
	MessageID string `json:"message_id"`
	MimeType  string `json:"mime_type"`
}

type ChatMessage struct {
	ID           string     `json:"id"`
	Message      string     `json:"message"`
	EmployeeID   int64      `json:"employee_id"`
	EmployeeName string     `json:"employee_name"`
	Files        []FileInfo `json:"files"`
	CreatedDt    string     `json:"created_dt"`
}

type ChatThread struct {
	Created       bool                   `json:"created"`
	ChatName      string                 `json:"chat_name"`
	EmployeeID    int64                  `json:"employee_id"`
	CreatedDt     string                 `json:"created_dt"`
	RootMessage   ChatMessage            `json:"root_message"`
	MessagesOrder []string               `json:"messages_order"`
	Messages      map[string]ChatMessage `json:"messages"`
}

type GetChatResponse struct {
	Chat           ChatThread `json:"chat"`
	HasMore        bool       `json:"has_more"`
	Limit          int        `json:"limit"`
	Before         string     `json:"before"`
	BeforeCreateAt string     `json:"before_create_at"`
}

type HistoryChatResponse struct {
	MessagesOrder  []string               `json:"messages_order"`
	RootMessage    ChatMessage            `json:"root_message"`
	Messages       map[string]ChatMessage `json:"messages"`
	HasMore        bool                   `json:"has_more"`
	Limit          int                    `json:"limit"`
	Before         string                 `json:"before"`
	BeforeCreateAt string                 `json:"before_create_at"`
}

type CreatePostResponse struct {
	TicketID int64       `json:"ticket_id"`
	Message  ChatMessage `json:"message"`
}

type GetDownloadLinkResponse struct {
	DownloadURL string `json:"download_url"`
}

const (
	WorkCategoriesApproveActionType = "status_approve_category"
	WorkCategoriesPerformActionType = "status_perform_category"
	WorkCategoriesViewActionType    = "view_ticket_category"
)

type TicketInfo struct {
	ChatID               string  `json:"chat_id"`
	TicketID             int64   `json:"ticket_id"`
	CategoryID           int64   `json:"category_id"`
	StatusID             string  `json:"status_id"`
	CreateEmployeeID     int64   `json:"create_employee_id"`
	GroupEmployeeIDs     []int64 `json:"group_employee_ids"`
	FavouriteEmployeeIDs []int64 `json:"favourite_employee_ids"`
}
