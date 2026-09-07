package chat_mapper

import (
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
)

type createdByEmployeeTickets struct {
	ActiveTickets    []ticketWithChat `json:"active_tickets"`
	RejectedTickets  []ticketWithChat `json:"rejected_tickets"`
	CompletedTickets []ticketWithChat `json:"completed_tickets"`
}

type ownTickets struct {
	ActiveTickets    []ticketWithChat `json:"active_tickets"`
	FavoriteTickets  []ticketWithChat `json:"favorite_tickets"`
	RejectedTickets  []ticketWithChat `json:"rejected_tickets"`
	CompletedTickets []ticketWithChat `json:"completed_tickets"`
}

type employeeWorkTickets struct {
	Approve       []ticketWithChat `json:"approve"`
	Perform       []ticketWithChat `json:"perform"`
	BookedPerform []ticketWithChat `json:"booked_perform"`
}

type ticketWithChat map[string]json.RawMessage

func (t ticketWithChat) getTicketIDAndChatExists() (int64, bool) {
	chatIDRaw, ok := t["chat_id"]
	if !ok || string(chatIDRaw) == "null" {
		return 0, false
	}

	var chatID string
	if err := jsoniter.Unmarshal(chatIDRaw, &chatID); err != nil || chatID == "" {
		return 0, false
	}

	var ticketID int64
	if err := jsoniter.Unmarshal(t["ticket_id"], &ticketID); err != nil || ticketID == 0 {
		return 0, false
	}

	return ticketID, true
}
func (t ticketWithChat) setChat(chat *models.ChatInfo) error {
	delete(t, "chat_id")
	if chat == nil {
		chat = &models.ChatInfo{Exists: false}
	}

	chatRaw, err := jsoniter.Marshal(chat)

	if err != nil {
		return fmt.Errorf("can't marshal chat: %w", err)
	}
	t["chat"] = chatRaw
	return nil
}
