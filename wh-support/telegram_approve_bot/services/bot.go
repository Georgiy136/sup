package services

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/cache"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/clients/auth_client"
	employeeInfo "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/clients/employee_info_api_client"
	whPortalExternalApiClient "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/clients/wh_portal_external_api_client"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type User struct {
	curTicketID int64
	state       int64
	extInfo     any
}

type Bot struct {
	Token             string `json:"token"`
	TimeoutRequestSec int    `json:"timeout_request_sec"`

	commandMap      map[string]commandInfo
	registrationMap sync.Map // map[int64]RegistrationExtInfo

	users sync.Map // map[tguser_id]User

	ticketRepo            ticketRepo
	userRepo              userRepo
	cacheRepo             cacheRepo
	ticketsCache          ticketsCache
	employeeInfoApiClient *employeeInfo.EmployeeInfoApiClient
	authorizationClient   *auth_client.AuthClient
	checkTelegramClient   *whPortalExternalApiClient.WhPortalExternalApiClient
	bot                   *tgbotapi.BotAPI
}

type userRepo interface {
	AddTgUser(chatID, employeeID int64) error
	GetEmpByUserID(userID int64) (int64, error)
}

type ticketRepo interface {
	TicketsApproveWithoutExt(ticketID, employeeID int64) error
	TicketsReject(ticketID, employeeID int64, comment string) error
	TicketsBook(ticketID, employeeID int64) error
	TicketsPerformWithoutAddInfo(ticketID int64, employeeID int64) error
	TicketsUnbook(ticketID int64, employeeID int64) error
}

type cacheRepo interface {
	SaveTelegramExistInfoWithBuildInTTL(userID int64, isExist bool) error
	GetTelegramExistInfo(userID int64) (*models.TelegramExistInfo, error)
}

type ticketsCache interface {
	AppendTicket(ticketID int64, msg models.TicketData) error
	GetTicketsByChatID(ticketID, chatID int64) ([]models.TicketData, error)
	Delete(ticketID, chatID int64) error
	DeleteByOperationType(ticketID, chatID int64, operationType string) error
}

func NewBot() *Bot {
	return &Bot{}
}

func (b *Bot) Init(ctx context.Context, config configs.Config) {
	const defaultTimeoutSec = 60

	logrus.Debug("initing bot")

	confRaw := config.GetByServiceKeyRequired("bot_config")
	err := jsoniter.Unmarshal(confRaw, b)
	if err != nil {
		logrus.Panicf("cannot unmarshal bot_config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		logrus.Panicf("cannot init bot: %v", err)
	}

	b.bot = bot
	logrus.Debugf("authorized on account %s", bot.Self.UserName)

	if b.TimeoutRequestSec <= 0 {
		b.TimeoutRequestSec = defaultTimeoutSec
		logrus.Debug("set default timeout for update")
	}

	b.ticketRepo = repository.NewTicketRepository()
	b.userRepo = repository.NewUserRepository()
	b.cacheRepo = cache.NewCache(config)
	b.ticketsCache = cache.NewTicketsCache(ctx, config)

	err = b.initCommands()
	if err != nil {
		logrus.Panicf("can't init commands: %v", err)
	}

	b.initRegistration(ctx, config)
	b.employeeInfoApiClient = employeeInfo.NewEmployeeInfoApiClient(ctx, config)
	b.checkTelegramClient = whPortalExternalApiClient.NewWhPortalExternalApiClient(ctx, config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		b.Stop()
	}()

	go b.botReply()
}

func (b *Bot) botReply() {
	u := tgbotapi.NewUpdate(0)
	u.AllowedUpdates = append(u.AllowedUpdates, tgbotapi.UpdateTypeMyChatMember)
	u.AllowedUpdates = append(u.AllowedUpdates, tgbotapi.UpdateTypeChatMember)
	u.AllowedUpdates = append(u.AllowedUpdates, tgbotapi.UpdateTypeMessage)
	u.AllowedUpdates = append(u.AllowedUpdates, tgbotapi.UpdateTypeInlineQuery)
	u.AllowedUpdates = append(u.AllowedUpdates, tgbotapi.UpdateTypeCallbackQuery)
	u.Timeout = b.TimeoutRequestSec
	updates := b.bot.GetUpdatesChan(u)

	logrus.Debug("starting reply cycle")

	for update := range updates {
		go func(update tgbotapi.Update) {
			err := b.handleUpdate(update)
			if err != nil {
				logrus.Errorf("can't handle update: %v", err)
			}
		}(update)
	}
}

func (b *Bot) Stop() {
	logrus.Debug("stopping bot")

	b.bot.StopReceivingUpdates()
}
