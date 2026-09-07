package services

import (
	"fmt"
	"strings"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	PhoneNumberLengthWithPlus    = 12
	PhoneNumberLengthWithoutPlus = 11
)

func replaceSpecSymbolsNotification(notification models.Notification) models.Notification {
	notification.TicketInfo.StatusId = utils.ReplaceSpecSymbols(notification.TicketInfo.StatusId)
	notification.TicketInfo.TicketName = utils.ReplaceSpecSymbols(notification.TicketInfo.TicketName)
	notification.TicketInfo.CategoryName = utils.ReplaceSpecSymbols(notification.TicketInfo.CategoryName)
	notification.TicketInfo.ParentCategoryName = utils.ReplaceSpecSymbols(notification.TicketInfo.ParentCategoryName)
	notification.TicketInfo.StatusDescription = utils.ReplaceSpecSymbols(notification.TicketInfo.StatusDescription)
	notification.TicketInfo.ScenarioName = utils.ReplaceSpecSymbols(notification.TicketInfo.ScenarioName)
	notification.TicketInfo.OperationType = utils.ReplaceSpecSymbols(notification.TicketInfo.OperationType)
	notification.Comments = utils.ReplaceSpecSymbols(notification.Comments)

	for idx := range notification.TicketInfo.Ext {
		notification.TicketInfo.Ext[idx].FrontDataName = utils.ReplaceSpecSymbols(notification.TicketInfo.Ext[idx].FrontDataName)
	}
	return notification
}

func getPhoneNumber(phoneNumber string) (string, error) {
	switch {
	case strings.HasPrefix(phoneNumber, "+7"):
		if len(phoneNumber) != PhoneNumberLengthWithPlus {
			return "", fmt.Errorf("get phone number: %w", common.ErrInvalidPlus7PhoneNumber)
		}
		return phoneNumber, nil

	case strings.HasPrefix(phoneNumber, "8"):
		if len(phoneNumber) != PhoneNumberLengthWithoutPlus {
			return "", fmt.Errorf("get phone number: %w", common.ErrInvalid8PhoneNumber)
		}
		phoneNumber = strings.Replace(phoneNumber, "8", "+7", 1)
		return phoneNumber, nil

	case strings.HasPrefix(phoneNumber, "7"):
		if len(phoneNumber) != PhoneNumberLengthWithoutPlus {
			return "", fmt.Errorf("get phone number: %w", common.ErrInvalid7PhoneNumber)
		}
		phoneNumber = strings.Replace(phoneNumber, "7", "+7", 1)
		return phoneNumber, nil

	default:
		return "", fmt.Errorf("get phone number: %w", common.ErrUndefinedPhoneNumberFormat)
	}
}

func (b *Bot) sendTryAgain(message *tgbotapi.Message) error {
	return b.replyToMsg(common.TryAgainLaterMsg, message, false, "", nil)
}

func (b *Bot) sendInvalidCode(message *tgbotapi.Message) error {
	const invalidCode = "Неверный код"
	return b.replyToMsg(invalidCode, message, false, "", nil)
}

func (b *Bot) sendDefault(text string, message *tgbotapi.Message) error {
	return b.replyToMsg(text, message, false, "Markdown", nil)
}

func (b *Bot) sendCustomParams(text string, message *tgbotapi.Message, disNotif bool, mode string, markup interface{}) error {
	return b.replyToMsg(text, message, disNotif, mode, markup)
}

func (b *Bot) sendAlreadyRegisteredWithoutReply(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, common.AlreadyRegisteredMsg)

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("can't send msg \"%s\": %w", common.AlreadyRegisteredMsg, err)
	}

	return nil
}

func (b *Bot) sendCustomMsg(chatID int64, msgText string) error {
	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = ParseModeMarkdown

	if _, err := b.sendMsg(msg); err != nil {
		if strings.Contains(err.Error(), "bot was blocked by the user") {
			return nil
		}
		return fmt.Errorf("can't send сustom msg: %w", err)
	}
	return nil
}

func (b *Bot) sendMessageWaitingComment(chatID int64, ticketID int64) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Опишите причину отклонения заявки *№ %d*:", ticketID))
	msg.ParseMode = ParseModeMarkdown

	if _, err := b.sendMsg(msg); err != nil {
		return fmt.Errorf("can't send msg: %w", err)
	}
	return nil
}

func (b *Bot) replyToMsg(msgToSend string, replyTo *tgbotapi.Message, disNotif bool, mode string, markup interface{}) error {
	if replyTo == nil {
		return fmt.Errorf("reply to msg: %w", common.ErrNilMsg)
	}

	msg := tgbotapi.NewMessage(replyTo.Chat.ID, msgToSend)
	msg.ReplyToMessageID = replyTo.MessageID
	msg.ParseMode = mode
	msg.DisableNotification = disNotif
	msg.ReplyMarkup = markup

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("can't reply: %w", err)
	}

	return nil
}

func getTicketsMsgByOperationType(ticketsMsgs []models.TicketData, operationType string) []models.TicketData {
	out := make([]models.TicketData, 0)

	for _, ticket := range ticketsMsgs {
		if ticket.Operation == operationType {
			out = append(out, ticket)
		}
	}

	return out
}

func (b *Bot) editTicketMessages(ticketMessages []models.TicketData, editText string, ticketID int64) {
	for idx := range ticketMessages {
		b.editTicketMessage(ticketMessages[idx], editText, ticketID)
	}
}
func (b *Bot) editTicketMessage(ticketMessage models.TicketData, editText string, ticketID int64) {
	editedMsg := tgbotapi.NewEditMessageText(
		ticketMessage.Message.ChatID,
		ticketMessage.Message.MessageID,
		ticketMessage.Message.Text+editText)
	editedMsg.ParseMode = ParseModeMarkdown

	if _, err := b.bot.Send(editedMsg); err != nil {
		if strings.Contains(err.Error(), "bot was blocked by the user") {
			return
		}
		logrus.Errorf("can't send edit message for ticket %d, operation type %s, chatID %d, can't send message with markdown: %v", ticketID, ticketMessage.Operation, ticketMessage.Message.ChatID, err)
		editedMsg.ParseMode = ""
		if _, err = b.bot.Send(editedMsg); err != nil {
			if strings.Contains(err.Error(), "bot was blocked by the user") {
				return
			}
			logrus.Errorf("can't send edit message for ticket %d, operation type %s, chatID %d, can't send message without parse mode: %v", ticketID, ticketMessage.Operation, ticketMessage.Message.ChatID, err)
		}
	}
}

func (b *Bot) sendMsg(msg tgbotapi.MessageConfig) (*tgbotapi.Message, error) {
	sentMsg, err := b.bot.Send(msg)
	if err != nil {
		if strings.Contains(err.Error(), "bot was blocked by the user") {
			return nil, nil
		}
		msg.ParseMode = ""
		sentMsg, err = b.bot.Send(msg)
		if err != nil {
			if strings.Contains(err.Error(), "bot was blocked by the user") {
				return nil, nil
			}
			return nil, fmt.Errorf("can't send message without parse mode: %v", err)
		}
	}
	return &sentMsg, nil
}

func (b *Bot) sendEditMsg(msg tgbotapi.EditMessageTextConfig) error {
	if _, err := b.bot.Send(msg); err != nil {
		if strings.Contains(err.Error(), "bot was blocked by the user") {
			return nil
		}
		msg.ParseMode = ""
		if _, err = b.bot.Send(msg); err != nil {
			if strings.Contains(err.Error(), "bot was blocked by the user") {
				return nil
			}
			return fmt.Errorf("can't send edit message without parse mode: %v", err)
		}
	}
	return nil
}
