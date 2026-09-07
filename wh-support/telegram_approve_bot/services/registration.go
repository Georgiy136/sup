package services

import (
	"context"
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/clients/auth_client"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type RegState int

const (
	WaitingNumber RegState = iota
	WaitingCode
)

type RegistrationExtInfo struct {
	state       RegState
	phoneNumber *string
}

func (b *Bot) initRegistration(ctx context.Context, config configs.Config) {
	b.authorizationClient = auth_client.New(fastclient.NewHttpClient())
	b.authorizationClient.Configure(ctx, config)
	b.registrationMap = sync.Map{}
}

func (b *Bot) registrationByPhoneNumberHandler(update tgbotapi.Update) error {
	markup := tgbotapi.NewOneTimeReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("Отправить телефон"),
		),
	)
	err := b.sendCustomParams("Пришлите пожалуйста ваш номер телефона\nдля дальнейшей работы", update.Message, false, "", markup)
	if err != nil {
		return fmt.Errorf("can't send start registration msg to tg: %w", err)
	}

	b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
		state: WaitingNumber,
	})

	return nil
}

func (b *Bot) registrationByWhPortalHandler(update tgbotapi.Update) error {
	err := b.checkCacheRegisteredUser(update.Message.From.ID)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrUserAlreadyRegistered):
			sErr := b.sendAlreadyRegisteredWithoutReply(update.FromChat().ID)
			if sErr != nil {
				return fmt.Errorf("can't send msg: %w", sErr)
			}
			return nil
		case errors.Is(err, common.ErrUserCannotRegistered):
			sErr := b.sendCustomMsg(update.FromChat().ID, common.TelegramNotFoundInWhPortal)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.TelegramNotFoundInWhPortal, sErr)
			}
			return nil
		default:
			return fmt.Errorf("can't check cache registered user: %w", err)
		}
	}

	resp, err := b.checkTelegramClient.CheckTelegramOnWhPortal(update.Message.From.ID)
	if err != nil || resp == nil {
		sErr := b.sendCustomMsg(update.FromChat().ID, common.UnexpectedErrorMsg)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"%s\": %w", common.UnexpectedErrorMsg, sErr)
		}

		return fmt.Errorf("can't check telegram on wh portal: %w", err)
	}

	err = b.cacheRepo.SaveTelegramExistInfoWithBuildInTTL(update.Message.From.ID, resp.IsExist)
	if err != nil {
		return fmt.Errorf("can't save info wh portal into cache: %w", err)
	}

	if !resp.IsExist {
		sErr := b.sendCustomMsg(update.FromChat().ID, common.TelegramNotFoundInWhPortal)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"%s\": %w", common.TelegramNotFoundInWhPortal, sErr)
		}
		return nil
	}

	err = b.registrationUser(resp.EmployeeID, update.Message.Chat.ID)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrChatAlreadyExists) || errors.Is(err, common.ErrUserAlreadyExists):
			sErr := b.sendAlreadyRegisteredWithoutReply(update.FromChat().ID)
			if sErr != nil {
				return fmt.Errorf("can't send msg: %w", sErr)
			}

			return nil
		case errors.Is(err, common.ErrHasNoRights):
			sErr := b.sendCustomMsg(update.FromChat().ID, common.HasNoRightsMsgError)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.HasNoRightsMsgError, sErr)
			}

			return nil
		default:
			sErr := b.sendCustomMsg(update.FromChat().ID, common.RegistrationMsgError)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.RegistrationMsgError, sErr)
			}

			return fmt.Errorf("can't registration user: %w", err)
		}
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Вы успешно зарегистрированы")
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("can't send msg: %w", err)
	}

	return nil
}

func (b *Bot) checkCacheRegisteredUser(userID int64) error {
	tgExistInfo, err := b.cacheRepo.GetTelegramExistInfo(userID)
	if err != nil {
		return fmt.Errorf("can't get data from cache: %w", err)
	}

	if tgExistInfo == nil {
		return nil
	}

	if tgExistInfo.IsExist {
		return fmt.Errorf("user %d is already registered: %w", userID, common.ErrUserAlreadyRegistered)
	}

	return fmt.Errorf("user %d can't be registered: %w", userID, common.ErrUserCannotRegistered)
}

func (b *Bot) handleRegistration(update tgbotapi.Update) error {
	var (
		userStateInfo RegistrationExtInfo
		tmp           any
		ok            bool
	)

	switch {
	case update.Message != nil:
		tmp, _ = b.registrationMap.Load(update.Message.From.ID)
	case update.CallbackQuery != nil:
		tmp, _ = b.registrationMap.Load(update.CallbackQuery.From.ID)
	default:
		return fmt.Errorf("handle registration: %w", common.ErrNorMsgNeitherCallback)
	}

	if userStateInfo, ok = tmp.(RegistrationExtInfo); !ok {
		sErr := b.sendTryAgain(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send msg: %w", sErr)
		}

		return fmt.Errorf("handle registration: %w", common.ErrInvalidType)
	}

	switch userStateInfo.state {
	case WaitingNumber:
		return b.scanNumber(update)
	case WaitingCode:
		return b.scanCode(update, userStateInfo.phoneNumber)
	default:
		return fmt.Errorf("can't define stat - %v: %w", userStateInfo.state, common.ErrUnknownState)
	}
}

func (b *Bot) scanNumber(update tgbotapi.Update) error {
	if update.Message == nil {
		return fmt.Errorf("wrong message format: %w", common.ErrNilMsg)
	}

	if update.Message.Contact == nil {
		markup := tgbotapi.NewOneTimeReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonContact("Отправить телефон"),
			),
		)
		sErr := b.sendCustomParams(common.UseContactButtonMsgText, update.Message, false, "", markup)
		if sErr != nil {
			return fmt.Errorf("can't send msg \"%s\": %w", common.UseContactButtonMsgText, sErr)
		}

		return fmt.Errorf("wrong format: %w", common.ErrContactMessage)
	}

	phoneNumber, err := getPhoneNumber(update.Message.Contact.PhoneNumber)
	if err != nil {
		sErr := b.sendDefault(common.InvalidNumberErrorMsg, update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"%s\": %w", common.InvalidNumberErrorMsg, sErr)
		}

		return fmt.Errorf("can't get phone number: %w", err)
	}

	err = b.authorizationClient.SendNotificationCode(phoneNumber)
	if err != nil {
		sErr := b.sendTryAgain(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send try again msg: %w", sErr)
		}

		return fmt.Errorf("can't send code: %w", err)
	}

	b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
		state:       WaitingCode,
		phoneNumber: &phoneNumber,
	})

	markup := tgbotapi.NewRemoveKeyboard(false)
	err = b.sendCustomParams("Пожалуйста пришлите код, отправленный на ваш аккаунт Wildberries", update.Message, false, "", markup)
	if err != nil {
		return fmt.Errorf("can't send code msg to tg: %w", err)
	}

	return nil
}

func (b *Bot) scanCode(update tgbotapi.Update, phoneNumber *string) error {
	if update.Message == nil {
		return fmt.Errorf("wrong message format: %w", common.ErrNilMsg)
	}

	code, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
			state: WaitingNumber,
		})

		sErr := b.sendInvalidCode(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"invalid code\": %w", sErr)
		}

		return fmt.Errorf("can't convert msg to code: %w", err)
	}

	if phoneNumber == nil {
		b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
			state: WaitingNumber,
		})

		sErr := b.sendInvalidCode(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send error msg \"invalid code\": %w", sErr)
		}

		return fmt.Errorf("invalid value: %w", common.ErrPhoneNumberNil)
	}

	employeeID, check, err := b.authorizationClient.CheckLogin(int64(code), *phoneNumber)
	if err != nil {
		b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
			state: WaitingNumber,
		})

		if errors.Is(err, common.ErrDeletedEmployee) {
			sErr := b.sendDefault(common.FiredEmployeeErrorMsg, update.Message)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.FiredEmployeeErrorMsg, sErr)
			}

			return fmt.Errorf("dismissed user: %w", err)
		}

		sErr := b.sendTryAgain(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send try again msg: %w", sErr)
		}

		return fmt.Errorf("can't check code: %w", err)
	}

	if !check {
		b.registrationMap.Store(update.Message.From.ID, RegistrationExtInfo{
			state: WaitingNumber,
		})

		sErr := b.sendTryAgain(update.Message)
		if sErr != nil {
			return fmt.Errorf("can't send try again msg: %w", sErr)
		}

		return fmt.Errorf("can't check login: %w", common.ErrCheckLogin)
	}

	b.users.Store(update.Message.From.ID, User{state: defaultState})

	err = b.registrationUser(employeeID, update.Message.Chat.ID)

	b.registrationMap.Delete(update.Message.From.ID)

	if err != nil {
		switch {
		case errors.Is(err, common.ErrChatAlreadyExists) || errors.Is(err, common.ErrUserAlreadyExists):
			sErr := b.sendAlreadyRegisteredWithoutReply(update.FromChat().ID)
			if sErr != nil {
				return fmt.Errorf("can't send msg: %w", sErr)
			}

			return nil
		case errors.Is(err, common.ErrHasNoRights):
			sErr := b.sendCustomMsg(update.FromChat().ID, common.HasNoRightsMsgError)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.HasNoRightsMsgError, sErr)
			}

			return nil
		default:
			sErr := b.sendCustomMsg(update.FromChat().ID, common.RegistrationMsgError)
			if sErr != nil {
				return fmt.Errorf("can't send error msg \"%s\": %w", common.RegistrationMsgError, sErr)
			}

			return fmt.Errorf("can't registration user: %w", err)
		}
	}

	msg := tgbotapi.NewMessage(update.Message.From.ID, "Вы успешно зарегистрированы")
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("can't send msg: %w", err)
	}

	return nil
}

func (b *Bot) registrationUser(employeeID int64, chatID int64) error {
	err := b.userRepo.AddTgUser(chatID, employeeID)
	if err != nil {
		return fmt.Errorf("can't add telegram user %w", err)
	}

	return nil
}
