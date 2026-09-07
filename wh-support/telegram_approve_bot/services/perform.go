package services

import (
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (b *Bot) genBookTicketMsg(msgText string, chatID, ticketID int64, typeOfEmployeeID string) tgbotapi.MessageConfig {
	message := tgbotapi.NewMessage(chatID, msgText)
	message.ReplyMarkup = b.genBookButton(ticketID, typeOfEmployeeID)
	message.ParseMode = ParseModeMarkdown

	return message
}

func (b *Bot) genBookButton(ticketID int64, typeOfEmployeeID string) tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	switch typeOfEmployeeID {
	case isGroup, isEmployee:
		row = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Взять в работу", fmt.Sprintf("%v %v", ticketID, book)),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("%v %v", ticketID, reject)),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(row)
}

func (b *Bot) genPerformButton(ticketID int64, typeOfEmployeeID string) tgbotapi.InlineKeyboardMarkup {
	switch typeOfEmployeeID {
	case isEmployee:
		performRow := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Исполнить", fmt.Sprintf("%v %v", ticketID, perform)),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("%v %v", ticketID, reject)),
		)

		unbookRow := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Снять с себя заявку", fmt.Sprintf("%v %v", ticketID, unbook)),
		)
		return tgbotapi.NewInlineKeyboardMarkup(performRow, unbookRow)
	}
	return tgbotapi.InlineKeyboardMarkup{}
}

func (b *Bot) performCallback(update tgbotapi.Update, ticketID int64, ticketMessages []models.TicketData) error {
	employeeID, err := b.userRepo.GetEmpByUserID(update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("can't get employeeID by tgUserID: %w", err)
	}

	if err = b.ticketRepo.TicketsPerformWithoutAddInfo(ticketID, employeeID); err != nil {
		if errors.Is(err, common.ErrTicketNotFound) {
			b.editTicketMessage(models.TicketData{Message: b.getMessageFromCallback(update)}, common.TicketNotFoundMessage, ticketID)
			return nil
		}
		return fmt.Errorf("can't ticket perform: %w", err)
	}

	for idx := range ticketMessages {
		editedMsg := tgbotapi.NewEditMessageText(
			ticketMessages[idx].Message.ChatID,
			ticketMessages[idx].Message.MessageID,
			ticketMessages[idx].Message.Text+b.getPerformedText(&employeeID))
		editedMsg.ParseMode = ParseModeMarkdown

		if err = b.sendEditMsg(editedMsg); err != nil {
			logrus.Errorf("can't send edit message on perform for ticket %d: %v", ticketID, err)
		}
	}

	if err = b.ticketsCache.Delete(ticketID, update.CallbackQuery.Message.Chat.ID); err != nil {
		logrus.Errorf("can't delete tickets: %v", err)
	}

	return nil
}

func (b *Bot) genPerformOnSiteTicketMsg(msgText string, chatID int64) tgbotapi.MessageConfig {
	message := tgbotapi.NewMessage(chatID, msgText+common.PerformOnSiteTicketMessage)
	message.ParseMode = ParseModeMarkdown

	return message
}

func (b *Bot) getPerformedText(performedEmployeeID *int64) string {
	if performedEmployeeID == nil {
		return "\n-------------\nЗаявка исполнена неизвестным пользователем ✅"
	}
	infoEmployee := b.getEmployeeNameByEmployeeID(*performedEmployeeID)

	return fmt.Sprintf("\n-------------\n%s (%d) исполнил(а) заявку ✅", infoEmployee, *performedEmployeeID)
}

func (b *Bot) handlePerformOperation(msgText string, notification models.Notification) error {
	var msg tgbotapi.MessageConfig
	if notification.TicketInfo.NeedAddedInfo == nil {
		return fmt.Errorf("need added info for gen perform message: %w", common.ErrNilValue)
	}

	if *notification.TicketInfo.NeedAddedInfo {
		msg = b.genPerformOnSiteTicketMsg(msgText, notification.ChatID)
	} else {
		msg = b.genBookTicketMsg(msgText, notification.ChatID, notification.TicketID, notification.TypeOfEmployeeID)
	}

	sentMsg, err := b.sendMsg(msg)
	if err != nil {
		return fmt.Errorf("can't send perform msg: %w", err)
	}

	if sentMsg == nil {
		return nil
	}

	if err = b.ticketsCache.AppendTicket(notification.TicketID, models.TicketData{
		Message: models.MessageInfo{
			ChatID:    sentMsg.Chat.ID,
			MessageID: sentMsg.MessageID,
			Text:      msgText,
		},
		Operation:        notification.TicketInfo.OperationType,
		TypeOfEmployeeID: notification.TypeOfEmployeeID,
	}); err != nil {
		return fmt.Errorf("can't append sent ticket: %w", err)
	}
	return nil
}

func (b *Bot) handlePerformedOperation(notification models.Notification) error {
	if notification.TypeOfEmployeeID == isCreator {
		if err := b.sendTicketMsgForCreator(notification); err != nil {
			return fmt.Errorf("send ticket for creator error: %w", err)
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
		b.getPerformedText(notification.TicketInfo.ResponsibleEmployeeID),
		notification.TicketID,
	)
	if err = b.ticketsCache.DeleteByOperationType(notification.TicketID, notification.ChatID, PerformOperation); err != nil {
		return fmt.Errorf("can't delete tickets msgs by operation type: %v", err)
	}
	return nil
}
