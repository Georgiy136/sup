package main

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/clients"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/services"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

func main() {
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()

	serviceWorker.Configuration(
		connection_control.InitConnections(env_config.GetCommonEnvConfigs()),
	)

	serviceWorker.Tasks(
		func(ctx context.Context, config configs.Config) {
			kafkaOlapClient := clients.NewWhKafkaOlapWebPublisherClient(ctx, config)

			getHandler := cron_core.GetWorkerRegistration(
				map[string]cron_core.Worker{
					"support.tickets.changes":           services.NewTicketsChangesExportCron(kafkaOlapClient),
					"support.tickets.categories":        services.NewTicketsCategoriesExportCron(kafkaOlapClient),
					"support.tickets.category.statuses": services.NewTicketsCategoryStatusesExportCron(kafkaOlapClient),
					"support.tickets.groups":            services.NewTicketsGroupsExportCron(kafkaOlapClient),
					"support.tickets.scenarios":         services.NewTicketsScenariosExportCron(kafkaOlapClient),
				},
			)
			getHandler.InitCrons(ctx, config)
		},
	)

	app.StartServer(serviceWorker)
}
