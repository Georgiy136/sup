package services

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
)

func (b *Bot) bookCallback(update tgbotapi.Update, ticketID int64, ticketMessages []models.TicketData) error {
	employeeID, err := b.userRepo.GetEmpByUserID(update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("can't get employeeID by tgUserID: %w", err)
	}

	if err = b.ticketRepo.TicketsBook(ticketID, employeeID); err != nil {
		if errors.Is(err, common.ErrTicketNotFound) {
			b.editTicketMessage(models.TicketData{Message: b.getMessageFromCallback(update)}, common.TicketNotFoundMessage, ticketID)
			return nil
		}
		return fmt.Errorf("can't booking ticket: %w", err)
	}

	for idx := range ticketMessages {
		if ticketMessages[idx].Message.MessageID != update.CallbackQuery.Message.MessageID {
			editedMsg := tgbotapi.NewEditMessageText(
				ticketMessages[idx].Message.ChatID,
				ticketMessages[idx].Message.MessageID,
				ticketMessages[idx].Message.Text+b.getBookedText(&employeeID))
			editedMsg.ParseMode = ParseModeMarkdown

			if err = b.sendEditMsg(editedMsg); err != nil {
				logrus.Errorf("can't send edit message on book for ticket %d: %v", ticketID, err)
			}
		}

		editedMsg := tgbotapi.NewEditMessageReplyMarkup(
			ticketMessages[idx].Message.ChatID,
			ticketMessages[idx].Message.MessageID,
			b.genPerformButton(ticketID, isEmployee))

		if _, err = b.bot.Send(editedMsg); err != nil {
			logrus.Errorf("can't send edit message with button on book for ticket %d: %v", ticketID, err)
		}
	}

	return nil
}

func (b *Bot) getBookedText(bookedEmployeeID *int64) string {
	if bookedEmployeeID == nil {
		return "\n-------------\nЗаявка была взята в работу неизвестным пользователем 💼"
	}
	infoEmployee := b.getEmployeeNameByEmployeeID(*bookedEmployeeID)

	return fmt.Sprintf("\n-------------\n%s (%d) взял(а) заявку в работу 💼", infoEmployee, *bookedEmployeeID)
}

func (b *Bot) handleBookedOperation(notification models.Notification) error {
	ticketMessages, err := b.ticketsCache.GetTicketsByChatID(notification.TicketID, notification.ChatID)
	if err != nil {
		return fmt.Errorf("can not get ticket %d: %w", notification.TicketID, err)
	}
	if len(ticketMessages) == 0 {
		return nil
	}

	b.bookedEditTicketMessages(ticketMessages, notification)
	return nil
}

func (b *Bot) bookedEditTicketMessages(ticketMessages []models.TicketData, notification models.Notification) {
	for idx := range ticketMessages {
		switch notification.TypeOfEmployeeID {
		case isEmployee:
			editedMsg := tgbotapi.NewEditMessageReplyMarkup(
				ticketMessages[idx].Message.ChatID,
				ticketMessages[idx].Message.MessageID,
				b.genPerformButton(notification.TicketID, notification.TypeOfEmployeeID))

			if _, err := b.bot.Send(editedMsg); err != nil {
				logrus.Errorf("can't send edit message with button on book for ticket %d: %v", notification.TicketID, err)
			}
		default:
			b.editTicketMessage(
				ticketMessages[idx],
				b.getBookedText(notification.TicketInfo.ResponsibleEmployeeID),
				notification.TicketID,
			)
		}
	}
}
