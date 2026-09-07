package services

import (
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"strconv"
	"strings"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type updateType uint

const (
	callback updateType = iota
	command
	messageText
	contact
	unknown
)

const (
	defaultState int64 = iota
	waitingRejectCommentState
)

const (
	approve int64 = iota
	book
	unbook
	perform
	reject
)

const (
	defaultBaseParseInt    = 10
	defaultBitSizeParseInt = 64
)

func (b *Bot) handleUpdate(update tgbotapi.Update) error {
	currentUpdateType, err := b.getUpdateType(update)
	if err != nil {
		return fmt.Errorf("can't get update type: %w", err)
	}
	switch currentUpdateType {
	case command:
		err := b.handleCommand(update)
		if err != nil {
			return fmt.Errorf("can't handle command: %w", err)
		}
	case messageText:
		err := b.handleMessageText(update)
		if err != nil {
			return fmt.Errorf("can't handle text message: %w", err)
		}
	case contact:
		err := b.handleContact(update)
		if err != nil {
			return fmt.Errorf("can't handle contact: %w", err)
		}
	case callback:
		err := b.handleCallback(update)
		if err != nil {
			return fmt.Errorf("can't handle callBack: %w", err)
		}
	case unknown:
		logMsg, err := jsoniter.MarshalToString(update)
		if err != nil {
			return fmt.Errorf("can't marshal update for log: %w", err)
		}
		return fmt.Errorf("unknown update type: %s", logMsg)
	default:
		return fmt.Errorf("can't process update type: %w", common.ErrUnknown)
	}
	return nil
}

func (b *Bot) handleCallback(update tgbotapi.Update) error {
	const validMinLenOfArgs = 2

	var (
		ticketMessages []models.TicketData
	)

	if update.CallbackQuery.Message == nil {
		return fmt.Errorf("invalid argyments: %w", common.ErrEmptyCallbackMessage)
	}

	argsSplit := strings.Split(update.CallbackQuery.Data, " ")
	if len(argsSplit) < validMinLenOfArgs {
		return fmt.Errorf("wrong callbacks settings: got %v", argsSplit)
	}

	ticketID, err := strconv.ParseInt(argsSplit[0], defaultBaseParseInt, defaultBitSizeParseInt)
	if err != nil {
		return fmt.Errorf("can't parse ticketID: %w", err)
	}

	action, err := strconv.ParseInt(argsSplit[1], defaultBaseParseInt, defaultBitSizeParseInt)
	if err != nil {
		return fmt.Errorf("can't parse action: %w", err)
	}

	if ticketMessages, err = b.ticketsCache.GetTicketsByChatID(ticketID, update.CallbackQuery.Message.Chat.ID); err != nil {
		return fmt.Errorf("can't get tickets: %w", err)
	}

	if len(ticketMessages) == 0 {
		ticketMessages = []models.TicketData{
			{
				Message:   b.getMessageFromCallback(update),
				Operation: b.getOperationByAction(action),
			},
		}
	}

	logrus.Infof("ticket %d button was pressed with action %d", ticketID, action)

	switch action {
	case approve:
		if err := b.approveCallback(update, ticketID, ticketMessages); err != nil {
			return fmt.Errorf("can't approve ticket %d: %w", ticketID, err)
		}
	case book:
		if err := b.bookCallback(update, ticketID, ticketMessages); err != nil {
			return fmt.Errorf("can't book ticket %d: %w", ticketID, err)
		}
	case unbook:
		if err := b.unbookCallback(update, ticketID, ticketMessages); err != nil {
			return fmt.Errorf("can't unbook ticket %d: %w", ticketID, err)
		}
	case perform:
		if err := b.performCallback(update, ticketID, ticketMessages); err != nil {
			return fmt.Errorf("can't perform ticket %d: %w", ticketID, err)
		}
	case reject:
		if err := b.rejectCallback(update, ticketID); err != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketID, err)
		}
	default:
		return common.ErrUnknownAction
	}
	return nil
}

func (b *Bot) handleCommand(update tgbotapi.Update) error {
	currCommand, exist := b.commandMap[update.Message.Command()]
	if !exist {
		sErr := b.sendCustomMsg(update.FromChat().ID, common.CommandNotExistErrorMsg)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"%s\": %w", common.CommandNotExistErrorMsg, sErr)
		}

		return nil
	}

	err := currCommand.handleFunc(update)
	if err != nil {
		return fmt.Errorf("can't handle command: %w", err)
	}
	return nil
}

func (b *Bot) handleMessageText(update tgbotapi.Update) error {
	if update.Message == nil {
		return fmt.Errorf("can't handle message text: %w", common.ErrExpectedNilTgMessage)
	}

	_, isRegistration := b.registrationMap.Load(update.Message.From.ID)
	if isRegistration {
		err := b.handleRegistration(update)
		if err != nil {
			return fmt.Errorf("can't handle registration: %w", err)
		}
		return nil
	}

	tmp, exist := b.users.Load(update.Message.From.ID)
	if !exist {
		return nil
	}

	user, ok := tmp.(User)
	if !ok {
		return fmt.Errorf("can't get user: %w", common.ErrInvalidType)
	}

	switch user.state {
	case waitingRejectCommentState:
		ticketMsg, ok := user.extInfo.(models.MessageInfo)
		if !ok {
			return fmt.Errorf("can't get extInfo for reject: %w", common.ErrInvalidType)
		}

		employeeID, err := b.userRepo.GetEmpByUserID(update.Message.From.ID)
		if err != nil {
			return fmt.Errorf("can't get employeeID by tgUserID: %w", err)
		}

		comment := update.Message.Text

		err = b.rejectTicket(update, user.curTicketID, employeeID, comment, ticketMsg)
		if err != nil {
			return fmt.Errorf("can't reject ticket %d: %w", user.curTicketID, err)
		}

		user.state = defaultState
		b.users.Store(update.Message.From.ID, user)
	case defaultState:
		return nil
	default:
		return fmt.Errorf("can't handle message text: %w", common.ErrInvalidState)
	}

	return nil
}

func (b *Bot) handleContact(update tgbotapi.Update) error {
	if update.Message == nil {
		return fmt.Errorf("can't handle message text: %w", common.ErrExpectedNilTgMessage)
	}

	_, isRegistration := b.registrationMap.Load(update.Message.From.ID)

	switch {
	case isRegistration:
		err := b.handleRegistration(update)
		if err != nil {
			return fmt.Errorf("can't handle registration: %w", err)
		}
	default:
		logrus.Warnf("can't handle contact message: %v", common.ErrInvalidState)
	}

	return nil
}

func (b *Bot) getUpdateType(update tgbotapi.Update) (updateType, error) {
	if update.CallbackQuery != nil {
		return callback, nil
	}

	if update.Message != nil {
		switch {
		case update.Message.IsCommand():
			return command, nil
		case update.Message.Contact != nil:
			return contact, nil
		case update.Message.Text != "":
			return messageText, nil
		}
	}

	return unknown, fmt.Errorf("can't get update type: %w", common.ErrUnknownUpdateType)
}

func (b *Bot) getMessageFromCallback(update tgbotapi.Update) models.MessageInfo {
	return models.MessageInfo{
		ChatID:    update.CallbackQuery.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.MessageID,
		Text:      update.CallbackQuery.Message.Text,
	}
}

func (b *Bot) getOperationByAction(action int64) string {
	switch action {
	case approve:
		return ApproveOperation
	case perform:
		return PerformOperation
	case reject:
		return RejectedOperation
	default:
		return ""
	}
}
