package models

import (
	"encoding/json"

	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
)

type BandSubmitDialogRequest struct {
	CallbackID string          `json:"callback_id" binding:"required"`
	State      string          `json:"state" binding:"required"`
	UserID     string          `json:"user_id" binding:"required"`
	ChannelID  string          `json:"channel_id" binding:"required"`
	Submission json.RawMessage `json:"submission"`
	Cancelled  bool            `json:"cancelled"`
}

type BandRejectSubmission struct {
	Comment string `json:"comment" binding:"required"`
}

type BandReturnSubmission struct {
	ReturnToStatusID string `json:"return_to_status_id"`
	Comment          string `json:"comment"`
}

type BandActionContext struct {
	TicketID         int64                       `json:"ticket_id" binding:"required,gt=0"`
	Action           string                      `json:"action" binding:"required"`
	Operation        string                      `json:"operation" binding:"required"`
	TypeOfEmployeeID ticketmodels.TypeOfEmployee `json:"type_of_employee_id" binding:"required"`
	EmployeeID       int64                       `json:"employee_id" binding:"required"`
	CategoryID       int64                       `json:"category_id"`
	StatusID         string                      `json:"status_id"`
	Signature        string                      `json:"sig" binding:"required"`
	Timestamp        int64                       `json:"ts" binding:"required"`
}

type BandActionRequest struct {
	UserID           string
	EmployeeID       int64
	ChannelID        string
	PostID           string
	TriggerID        string
	Comment          string
	TicketID         int64
	CategoryID       int64
	Action           string
	Operation        string
	TypeOfEmployeeID ticketmodels.TypeOfEmployee
	StatusID         string
	ReturnToStatusID string
}

type BandDialogState struct {
	TicketID         int64                       `json:"ticket_id"`
	CategoryID       int64                       `json:"category_id"`
	StatusID         string                      `json:"status_id"`
	Action           string                      `json:"action"`
	Operation        string                      `json:"operation"`
	TypeOfEmployeeID ticketmodels.TypeOfEmployee `json:"type_of_employee_id"`
	EmployeeID       int64                       `json:"employee_id"`
	ChannelID        string                      `json:"channel_id"`
	PostID           string                      `json:"post_id"`
	Signature        string                      `json:"sig"`
	Timestamp        int64                       `json:"ts"`
}

type BandSentNotification struct {
	PostID    string
	ChannelID string
}

type BandRecipient struct {
	EmployeeID int64
}
