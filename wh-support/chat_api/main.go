package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/repository"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/service"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"

	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
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
	runner.EnableAuthActionsCheck()

	redisClient := clients.NewRedisClient()
	chatProviderApiClient := clients.NewChatProviderApiClient()
	resourceEmployeeAccessClient := clients.NewResourceEmployeeAccessClient()
	employeeInfoApiClient := clients.NewEmployeeInfoApiClient()
	repo := repository.NewTicketsRepo()

	runner.RegisterConfigurableEntity(redisClient.Configure)
	runner.RegisterConfigurableEntity(chatProviderApiClient.Configure)
	runner.RegisterConfigurableEntity(resourceEmployeeAccessClient.Configure)
	runner.RegisterConfigurableEntity(employeeInfoApiClient.Configure)

	srvChat := service.NewChatService(repo, redisClient, chatProviderApiClient, resourceEmployeeAccessClient, employeeInfoApiClient)

	runner.RegisterCustomHandler("GetChat", srvChat.GetChat)
	runner.RegisterCustomHandler("CreatePost", srvChat.CreatePost)
	runner.RegisterCustomHandler("GetHistoryChat", srvChat.GetHistoryChat)
	runner.RegisterCustomHandler("ViewChatByUser", srvChat.ViewChatByUser)
	runner.RegisterCustomHandler("GetTicketEmployeesForChatSelector", srvChat.GetTicketEmployeesForChatSelector)

	runner.Run(loggerToClickhouse...)
}
