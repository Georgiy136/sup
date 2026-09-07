package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/personal_account_api/internal/service"
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

	srv := service.NewPersonalAccount()

	runner.RegisterConfigurableEntity(srv.Configure)

	runner.RegisterCustomHandler("GetPersonalAccountInfoV2", srv.GetPersonalAccountInfoV2)
	runner.RegisterCustomHandler("GetGroupsByEmployee", srv.GetGroupsByEmployee)

	runner.Run(loggerToClickhouse...)
}
