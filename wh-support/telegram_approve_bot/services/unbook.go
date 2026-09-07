package services

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"strings"
)

func (b *Bot) unbookCallback(update tgbotapi.Update, ticketID int64, ticketMessages []models.TicketData) error {
	employeeID, err := b.userRepo.GetEmpByUserID(update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("can't get employeeID by tgUserID: %w", err)
	}

	if err = b.ticketRepo.TicketsUnbook(ticketID, employeeID); err != nil {
		if errors.Is(err, common.ErrTicketNotFound) {
			b.editTicketMessage(models.TicketData{Message: b.getMessageFromCallback(update)}, common.TicketNotFoundMessage, ticketID)
			return nil
		}
		return fmt.Errorf("can't unbooking ticket: %w", err)
	}

	for idx := range ticketMessages {
		if ticketMessages[idx].Message.MessageID != update.CallbackQuery.Message.MessageID {
			newMessage := b.getUnbookedText(ticketMessages[idx].Message.Text, &employeeID)

			editedMsg := tgbotapi.NewEditMessageText(
				ticketMessages[idx].Message.ChatID,
				ticketMessages[idx].Message.MessageID,
				newMessage)
			editedMsg.ParseMode = ParseModeMarkdown

			if err = b.sendEditMsg(editedMsg); err != nil {
				logrus.Errorf("can't send edit message on unbook for ticket %d: %v", ticketID, err)
			}
		}

		editedButtonMsg := tgbotapi.NewEditMessageReplyMarkup(
			ticketMessages[idx].Message.ChatID,
			ticketMessages[idx].Message.MessageID,
			b.genBookButton(ticketID, ticketMessages[idx].TypeOfEmployeeID))

		if _, err = b.bot.Send(editedButtonMsg); err != nil {
			logrus.Errorf("can't send edit message with button on unbook for ticket %d: %v", ticketID, err)
		}
	}

	return nil
}

func (b *Bot) unbookEditTicketMessages(ticketMessages []models.TicketData, ticketID int64, responsibleEmployeeID *int64) {
	for idx := range ticketMessages {
		newMessage := b.getUnbookedText(ticketMessages[idx].Message.Text, responsibleEmployeeID)

		editedMsg := tgbotapi.NewEditMessageText(
			ticketMessages[idx].Message.ChatID,
			ticketMessages[idx].Message.MessageID,
			newMessage)
		editedMsg.ParseMode = ParseModeMarkdown

		if err := b.sendEditMsg(editedMsg); err != nil {
			logrus.Errorf("can't send edit message on unbook for ticket %d: %v", ticketID, err)
		}

		editedButtonMsg := tgbotapi.NewEditMessageReplyMarkup(
			ticketMessages[idx].Message.ChatID,
			ticketMessages[idx].Message.MessageID,
			b.genBookButton(ticketID, ticketMessages[idx].TypeOfEmployeeID))

		if _, err := b.bot.Send(editedButtonMsg); err != nil {
			logrus.Errorf("can't send edit message with button on unbook for ticket %d: %v", ticketID, err)
		}
	}
}

func (b *Bot) getUnbookedText(text string, unBookedEmployeeID *int64) string {
	var msgCut string
	if unBookedEmployeeID == nil {
		msgCut = "\n-------------\nЗаявка была взята в работу неизвестным пользователем 💼"
	} else {
		infoEmployee := b.getEmployeeNameByEmployeeID(*unBookedEmployeeID)
		msgCut = fmt.Sprintf("\n-------------\n%s (%d) взял(а) заявку в работу 💼", infoEmployee, *unBookedEmployeeID)
	}

	if before, _, found := strings.Cut(text, msgCut); found {
		return before
	}
	return text
}

func (b *Bot) handleUnBookedOperation(notification models.Notification) error {
	ticketMessages, err := b.ticketsCache.GetTicketsByChatID(notification.TicketID, notification.ChatID)
	if err != nil {
		return fmt.Errorf("can not get ticket %d: %w", notification.TicketID, err)
	}
	if len(ticketMessages) == 0 {
		return nil
	}

	b.unbookEditTicketMessages(
		ticketMessages,
		notification.TicketID,
		notification.TicketInfo.ResponsibleEmployeeID,
	)
	return nil
}
