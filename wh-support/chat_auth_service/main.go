package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/jwt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/service"
	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	runner := rest_auto_api.NewService()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	runner.RegisterConfigurableEntity(
		connection_control.InitConnections(conf),
	)

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	jwtGenerator := jwt.NewJwtTokenGenerator()
	redisClient := clients.NewRedisClient()
	resourceEmployeeAccessClient := clients.NewResourceEmployeeAccessClient()

	runner.RegisterConfigurableEntity(jwtGenerator.Configure)
	runner.RegisterConfigurableEntity(redisClient.Configure)
	runner.RegisterConfigurableEntity(resourceEmployeeAccessClient.Configure)

	srv := service.NewChatAuthService(jwtGenerator, redisClient, resourceEmployeeAccessClient)

	runner.RegisterConfigurableEntity(srv.Configure)
	runner.RegisterCustomHandler("GetRealtimeToken", srv.GetRealtimeToken)
	runner.RegisterCustomHandler("ChatAuthSubscribe", srv.ChatAuthSubscribe)
	runner.RegisterCustomHandler("ChatAuthRefresh", srv.ChatAuthRefresh)

	runner.Run(loggerToClickhouse...)
}
