package models

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type CentrifugoPublishRequest struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

type CentrifugoMsgPublishData struct {
	Event        string     `json:"event"`
	Message      string     `json:"message"`
	EmployeeID   int64      `json:"employee_id"`
	EmployeeName string     `json:"employee_name"`
	CreatedDt    string     `json:"created_dt"`
	Files        []FileInfo `json:"files"`
}

type CentrifugoUserNotifyPublishData struct {
	Event    string `json:"event"`
	TicketID int64  `json:"ticket_id"`
	Notify   Notify `json:"notify"`
	Meta     Meta   `json:"meta"`
}

type CentrifugoChatNotificationPublishData struct {
	Event              string  `json:"event"`
	TicketID           int64   `json:"ticket_id"`
	Message            string  `json:"message"`
	MessageID          string  `json:"message_id"`
	ChatID             string  `json:"chat_id"`
	SenderEmployeeID   int64   `json:"sender_employee_id"`
	SenderEmployeeName string  `json:"sender_employee_name"`
	NotifyEmployeeIDs  []int64 `json:"notify_employee_ids"`
	CreatedDt          string  `json:"created_dt"`
}

type Notify struct {
	Exists        bool   `json:"exists"`
	LastMessageId string `json:"last_message_id"`
	LastMessageDt string `json:"last_message_dt"`
}
type Meta struct {
	PublishedAt time.Time `json:"published_at"`
}

type CentrifugoResponse struct {
	Result jsoniter.RawMessage `json:"result"`
	Error  *CentrifugoError    `json:"error"`
}

type CentrifugoError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
