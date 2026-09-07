package services

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
)

func (b *Bot) rejectCallback(update tgbotapi.Update, ticketID int64) error {
	var (
		user User
		ok   bool
	)

	tmp, exist := b.users.Load(update.CallbackQuery.From.ID)
	if !exist {
		user = User{state: defaultState}
	} else {
		user, ok = tmp.(User)
		if !ok {
			return fmt.Errorf("can't get user state: %w", common.ErrInvalidType)
		}
	}

	user.curTicketID = ticketID
	user.state = waitingRejectCommentState
	user.extInfo = models.MessageInfo{ChatID: update.CallbackQuery.From.ID, MessageID: update.CallbackQuery.Message.MessageID, Text: update.CallbackQuery.Message.Text}

	b.users.Store(update.CallbackQuery.From.ID, user)

	err := b.sendMessageWaitingComment(update.CallbackQuery.Message.Chat.ID, ticketID)
	if err != nil {
		return fmt.Errorf("can't send message waiting comment: %w", err)
	}

	return nil
}

func (b *Bot) rejectTicket(update tgbotapi.Update, ticketID, employeeID int64, comment string, ticketMsg models.MessageInfo) error {
	if err := b.ticketRepo.TicketsReject(ticketID, employeeID, comment); err != nil {
		if errors.Is(err, common.ErrTicketNotFound) {
			b.editTicketMessage(models.TicketData{Message: ticketMsg}, common.TicketNotFoundMessage, ticketID)
			return nil
		}
		return fmt.Errorf("can't ticket reject: %w", err)
	}

	ticketMessages, err := b.ticketsCache.GetTicketsByChatID(ticketID, update.FromChat().ID)
	if err != nil {
		return fmt.Errorf("can't get tickets: %w", err)
	}

	if len(ticketMessages) == 0 {
		if err = b.sendCustomMsg(update.FromChat().ID, b.getRejectText(&employeeID)); err != nil {
			return fmt.Errorf("can't send custom message for ticket %d: %v", ticketID, err)
		}
		return nil
	}

	b.editTicketMessages(
		ticketMessages,
		b.getRejectText(&employeeID),
		ticketID,
	)

	if err = b.ticketsCache.Delete(ticketID, update.FromChat().ID); err != nil {
		logrus.Errorf("can't delete tickets: %v", err)
	}

	return nil
}

func (b *Bot) genRejectTicketMsg(chatID int64, msgText string, rejectedEmployeeID *int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, msgText+b.getRejectText(rejectedEmployeeID))
	msg.ParseMode = ParseModeMarkdown
	return msg
}

func (b *Bot) getRejectText(rejectEmployeeID *int64) string {
	if rejectEmployeeID == nil {
		return "\n-------------\nЗаявка была отклонена неизвестным пользователем ❌"
	}
	infoEmployee := b.getEmployeeNameByEmployeeID(*rejectEmployeeID)

	return fmt.Sprintf("\n-------------\n%s (%d) отклонил(а) заявку ❌", infoEmployee, *rejectEmployeeID)
}

func (b *Bot) handleRejectedOperation(msgText string, notification models.Notification) error {
	if notification.TypeOfEmployeeID == isCreator {
		if err := b.sendTicketMsgForCreator(notification); err != nil {
			return fmt.Errorf("send ticket for creator error: %w", err)
		}
	} else {
		msg := b.genRejectTicketMsg(notification.ChatID, msgText, notification.TicketInfo.RejectedEmployeeID)
		if _, err := b.sendMsg(msg); err != nil {
			return fmt.Errorf("can't send reject msg: %w", err)
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
		b.getRejectText(notification.TicketInfo.RejectedEmployeeID),
		notification.TicketID,
	)
	if err = b.ticketsCache.Delete(notification.TicketID, notification.ChatID); err != nil {
		return fmt.Errorf("can't delete tickets: %v", err)
	}
	return nil
}
