package services

import (
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (b *Bot) genApproveTicketMsg(msgText string, chatID, ticketID int64, typeOfEmployeeID string) tgbotapi.MessageConfig {
	row := []tgbotapi.InlineKeyboardButton{}
	switch typeOfEmployeeID {
	case isGroup, isEmployee:
		row = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", fmt.Sprintf("%v %v", ticketID, approve)),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("%v %v", ticketID, reject)),
		)
	}

	message := tgbotapi.NewMessage(chatID, msgText)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)
	message.ParseMode = ParseModeMarkdown

	return message
}

func (b *Bot) approveCallback(update tgbotapi.Update, ticketID int64, ticketMessages []models.TicketData) error {
	employeeID, err := b.userRepo.GetEmpByUserID(update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("can't get employeeID by tgUserID: %w", err)
	}

	addText := b.getApproveText(&employeeID)

	if err = b.ticketRepo.TicketsApproveWithoutExt(ticketID, employeeID); err != nil {
		ticketMessage := models.TicketData{Message: b.getMessageFromCallback(update)}
		switch {
		case errors.Is(err, common.ErrWrongTicketModel):
			b.editTicketMessage(ticketMessage, common.ApproveOnSiteTicketMessage, ticketID)
			return nil
		case errors.Is(err, common.ErrTicketNotNeedApprove):
			b.editTicketMessage(ticketMessage, common.TicketNotNeedApproveMessage, ticketID)
			return nil
		case errors.Is(err, common.ErrTicketNotFound):
			b.editTicketMessage(ticketMessage, common.TicketNotFoundMessage, ticketID)
			return nil
		default:
			return fmt.Errorf("can't approve ticket № %d: %w", ticketID, err)
		}
	}

	for idx := range ticketMessages {
		if ticketMessages[idx].Operation == ApproveOperation {
			editedMsg := tgbotapi.NewEditMessageText(
				ticketMessages[idx].Message.ChatID,
				ticketMessages[idx].Message.MessageID,
				ticketMessages[idx].Message.Text+addText)
			editedMsg.ParseMode = ParseModeMarkdown

			if err = b.sendEditMsg(editedMsg); err != nil {
				logrus.Errorf("can't send edit message on approve for ticket %d: %v", ticketID, err)
			}
		}
	}

	if err = b.ticketsCache.Delete(ticketID, update.CallbackQuery.Message.Chat.ID); err != nil {
		logrus.Errorf("can't delete tickets: %v", err)
	}

	return nil
}

func (b *Bot) genApproveOnSiteTicketMsg(msgText string, chatID int64) tgbotapi.MessageConfig {
	message := tgbotapi.NewMessage(chatID, msgText+common.ApproveOnSiteTicketMessage)
	message.ParseMode = ParseModeMarkdown

	return message
}

func (b *Bot) getApproveText(approveEmployeeID *int64) string {
	if approveEmployeeID == nil {
		return "\n-------------\nЗаявка была подтверждена неизвестным пользователем ✅"
	}
	infoEmployee := b.getEmployeeNameByEmployeeID(*approveEmployeeID)

	return fmt.Sprintf("\n-------------\n%s (%d) подтвердил(а) заявку ✅", infoEmployee, *approveEmployeeID)
}

func (b *Bot) handleApproveOperation(msgText string, notification models.Notification) error {
	var msg tgbotapi.MessageConfig
	if notification.TicketInfo.NeedAddedInfo != nil && *notification.TicketInfo.NeedAddedInfo {
		msg = b.genApproveOnSiteTicketMsg(msgText, notification.ChatID)
	} else {
		msg = b.genApproveTicketMsg(msgText, notification.ChatID, notification.TicketID, notification.TypeOfEmployeeID)
	}

	sentMsg, err := b.sendMsg(msg)
	if err != nil {
		return fmt.Errorf("can't send approve msg: %w", err)
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

func (b *Bot) handleApprovedOperation(notification models.Notification) error {
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
		getTicketsMsgByOperationType(ticketMessages, ApproveOperation),
		b.getApproveText(notification.TicketInfo.ResponsibleEmployeeID),
		notification.TicketID,
	)

	if err = b.ticketsCache.DeleteByOperationType(notification.TicketID, notification.ChatID, ApproveOperation); err != nil {
		return fmt.Errorf("can't delete tickets msgs by operation type: %v", err)
	}
	return nil
}
