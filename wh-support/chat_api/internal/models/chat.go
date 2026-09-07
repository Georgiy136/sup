package models

type ViewChatByUserResponse struct {
	TicketID     int64  `json:"ticket_id"`
	TimeViewChat string `json:"time_view_chat"`
}
