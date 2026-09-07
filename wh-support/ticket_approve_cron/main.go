package main

import (
	"context"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/clients"
	repo "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/repository"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/services"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()
	serviceWorker.Configuration(connection_control.InitConnections(conf))

	approveTelegramClient := clients.NewTelegramApproveBotClient()
	ticketRepo := repo.NewTicketsRepo()
	logID := cron_core.NewInMemoryLogID()

	serviceWorker.Tasks(func(ctx context.Context, conf configs.Config) {
		worker := cron_core.HandlerRegistration{Workers: map[string]cron_core.Worker{
			"ticket_approve_cron": services.NewTicketApproveCron(approveTelegramClient, ticketRepo, logID),
		}}
		worker.InitCrons(ctx, conf)
	})

	app.StartServer(serviceWorker)
}
