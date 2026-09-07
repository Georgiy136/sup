package main

import (
	"context"

	bandaction "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/band_action"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/bots"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/crons"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/handlers"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/mode"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat/events"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/signer"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/storage/db"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/storage/memory"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/storage/redis"
	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableBasicWithHmacAuth()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	ticketActionsRepo := db.NewTicketActionsRepo()
	notificationsRepo := db.NewNotificationsRepo()
	employeesRepo := db.NewEmployeesRepo()
	employeeInfoApiClient := clients.NewEmployeeInfoApiClient()
	inMemoryCache := memory.NewInMemoryCache()
	bandBotSender := bots.NewBandBot(inMemoryCache)
	centrifugoClient := clients.NewCentrifugoClient()
	cache := redis.NewStorage()
	actionSigner := signer.New()
	modeChecker := mode.NewModeChecker(employeesRepo)

	runner.RegisterConfigurableEntity(connection_control.InitConnections(conf))
	runner.RegisterConfigurableEntity(employeeInfoApiClient.Configure)
	runner.RegisterConfigurableEntity(inMemoryCache.Configure)
	runner.RegisterConfigurableEntity(bandBotSender.Configure)
	runner.RegisterConfigurableEntity(centrifugoClient.Configure)
	runner.RegisterConfigurableEntity(cache.Configure)
	runner.RegisterConfigurableEntity(actionSigner.Configure)
	runner.RegisterConfigurableEntity(modeChecker.Configure)

	bandActionService := bandaction.NewBandActionService(
		ticketActionsRepo,
		bandBotSender,
		cache,
		actionSigner,
	)
	ticketNotificationsService := ticket.NewTicketNotificationsService(
		ticket.NewNotificationBuilder(employeeInfoApiClient),
		ticketActionsRepo,
		bandBotSender,
		cache,
		notificationsRepo,
		modeChecker,
		bandActionService,
	)
	bandActionHandler := handlers.NewBandActionHandler(bandActionService, actionSigner)

	chatNotificationsService := chat.NewChatNotificationsService(
		centrifugoClient,
		cache,
		events.NewChatMessageEventNotifier(centrifugoClient, cache, bandBotSender, modeChecker),
	)

	runner.RegisterCustomHandler("BandTicketAction", bandActionHandler.HandleBandAction)
	runner.RegisterCustomHandler("BandRejectDialog", bandActionHandler.HandleBandRejectDialog)
	runner.RegisterCustomHandler("BandAdditionalInfoDialog", bandActionHandler.HandleBandAdditionalInfoDialog)
	runner.RegisterCustomHandler("BandReturnToStatusDialog", bandActionHandler.HandleBandReturnToStatusDialog)

	runner.RegisterServerStartTasks(func(ctx context.Context, conf configs.Config) {
		worker := cron_core.HandlerRegistration{Workers: map[string]cron_core.Worker{
			"ticket_notifications_cron": crons.NewSendNotificationsCron(ticketNotificationsService),
			"chat_notifications_cron":   crons.NewSendNotificationsCron(chatNotificationsService),
		}}
		worker.InitCrons(ctx, conf)
	})

	runner.Run(loggerToClickhouse...)
}
