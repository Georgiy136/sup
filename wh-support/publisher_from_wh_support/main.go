package main

import (
	"context"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

func main() {
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()
	conf := env_config.GetCommonEnvConfigs()

	serviceWorker.Configuration(connection_control.InitConnections(conf))

	serviceWorker.Tasks(func(ctx context.Context, config configs.Config) {
		//publisher_by_timer.NewTimerPublisher(map[string]publisher_by_timer.Handler{
		//	"tickets.status.change.publisher": new(service.TicketsStatusChanges),
		//}).Start(ctx, config)
	})

	app.StartServer(serviceWorker)
}
