package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/statistics_api/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/statistics_api/internal/service"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"

	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	runner := rest_auto_api.NewService()

	runner.RegisterConfigurableEntity(
		connection_control.InitConnections(conf),
	)
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	redisClient := clients.NewRedisClient()
	resourceEmployeeAccessClient := clients.NewResourceEmployeeAccessClient()

	runner.RegisterConfigurableEntity(redisClient.Configure)
	runner.RegisterConfigurableEntity(resourceEmployeeAccessClient.Configure)

	srvStatistics := service.NewStatistics(redisClient, resourceEmployeeAccessClient)
	runner.RegisterCustomHandler("GetStatisticsByEmployeeV2", srvStatistics.GetStatisticsByEmployeeV2)
	runner.RegisterCustomHandler("GetTicketStatistics", srvStatistics.GetTicketStatistics)

	runner.Run(loggerToClickhouse...)
}
