package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
)

func (b *Bot) genCompletedTicketMsg(chatID int64, msgText string, responsibleEmployeeID *int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, msgText+b.getCompletedText(responsibleEmployeeID))
	msg.ParseMode = ParseModeMarkdown
	return msg
}

func (b *Bot) getCompletedText(completedEmployeeID *int64) string {
	if completedEmployeeID == nil {
		return "\n-------------\nЗаявка выполнена неизвестным пользователем ✅"
	}
	infoEmployee := b.getEmployeeNameByEmployeeID(*completedEmployeeID)

	return fmt.Sprintf("\n-------------\n%s (%d) выполнил(а) заявку ✅", infoEmployee, *completedEmployeeID)
}

func (b *Bot) handleCompletedOperation(msgText string, notification models.Notification) error {
	if notification.TypeOfEmployeeID == isCreator {
		if err := b.sendTicketMsgForCreator(notification); err != nil {
			return fmt.Errorf("send ticket for creator error: %w", err)
		}
	} else {
		msg := b.genCompletedTicketMsg(notification.ChatID, msgText, notification.TicketInfo.ResponsibleEmployeeID)
		if _, err := b.sendMsg(msg); err != nil {
			return fmt.Errorf("can't send completed msg: %w", err)
		}
	}

	ticketMessages, err := b.ticketsCache.GetTicketsByChatID(notification.TicketID, notification.ChatID)
	if err != nil {
		return fmt.Errorf("can not get ticket %d: %w", notification.TicketID, err)
	}
	if len(ticketMessages) == 0 {
		return nil
	}

	b.editTicketMessages(
		ticketMessages,
		b.getCompletedText(notification.TicketInfo.ResponsibleEmployeeID),
		notification.TicketID,
	)
	if err = b.ticketsCache.Delete(notification.TicketID, notification.ChatID); err != nil {
		return fmt.Errorf("can't delete tickets: %v", err)
	}
	return nil
}
