package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type commandInfo struct {
	command    tgbotapi.BotCommand
	handleFunc func(update tgbotapi.Update) error
}

const startMessage = `Вас приветствует @%v 👋
Я умею:
- Рассылать уведомления о новых заявках

Перед использованием бота - в нем нужно зарегистрироваться, используйте команды:
/register_by_phone_number - для регистрации по номеру телефона
/register_by_wh_portal - для регистрации через WH Portal`

func (b *Bot) initCommands() error {
	b.commandMap = map[string]commandInfo{
		"start": {
			command:    tgbotapi.BotCommand{Command: "start", Description: "Приветственное сообщение"},
			handleFunc: b.startCommandHandler,
		},
		"register_by_phone_number": {
			command:    tgbotapi.BotCommand{Command: "register_by_phone_number", Description: "Зарегистрироваться в боте по номеру телефона"},
			handleFunc: b.registrationByPhoneNumberHandler,
		},
		"register_by_wh_portal": {
			command:    tgbotapi.BotCommand{Command: "register_by_wh_portal", Description: "Зарегистрироваться в боте через Wh Portal"},
			handleFunc: b.registrationByWhPortalHandler,
		},
	}

	err := b.setCommands()
	if err != nil {
		return fmt.Errorf("can't set commands: %v", err)
	}

	return nil
}

func (b *Bot) setCommands() error {
	logrus.Debug("setting bot commands")

	commands := make([]tgbotapi.BotCommand, 0, len(b.commandMap))
	for _, v := range b.commandMap {
		commands = append(commands, v.command)
	}

	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.bot.Request(setCommandsConfig)
	if err != nil {
		return fmt.Errorf("can't set commands to bot: %w", err)
	}

	return nil
}

func (b *Bot) startCommandHandler(update tgbotapi.Update) error {
	logrus.Debug("called /start command")

	err := b.sendCustomMsg(update.Message.Chat.ID, fmt.Sprintf(startMessage, b.bot.Self.UserName))
	if err != nil {
		return fmt.Errorf("can't send start msg to tg: %w", err)
	}

	return nil
}
