package models

type CreateNewChatRequest struct {
	TicketID int64 `json:"ticket_id"`
}

type CreateNewChatResponse struct {
	TicketID int64 `json:"ticket_id"`
	Chat     Chat  `json:"chat"`
}

type CreatePostRequest struct {
	ChatID     string   `json:"chat_id"`
	MessageStr string   `json:"message"`
	TicketID   int64    `json:"ticket_id"`
	FileIDs    []string `json:"file_ids"`
}

type CreatePostResponse struct {
	TicketID int64       `json:"ticket_id"`
	Message  ChatMessage `json:"message"`
}

type GetChatByChatIDRequest struct {
	ChatID string `json:"chat_id"`
}

type GetChatByChatIDResponse struct {
	Chat           Chat   `json:"chat"`
	HasMore        bool   `json:"has_more"`
	Limit          int64  `json:"limit"`
	Before         string `json:"before"`
	BeforeCreateAt string `json:"before_create_at"`
}

type GetHistoryChatByChatIDRequest struct {
	ChatID       string `json:"chat_id"`
	CountPost    int64  `json:"count_post"`
	FromPost     string `json:"from_post"`
	FromCreateAt string `json:"from_create_at"`
}
type GetHistoryChatByChatIDResponse struct {
	MessagesOrder  []string               `json:"messages_order"`
	ChatMessages   map[string]ChatMessage `json:"messages"`
	RootMessage    ChatMessage            `json:"root_message"`
	HasMore        bool                   `json:"has_more"`
	Limit          int64                  `json:"limit"`
	Before         string                 `json:"before"`
	BeforeCreateAt string                 `json:"before_create_at"`
}
