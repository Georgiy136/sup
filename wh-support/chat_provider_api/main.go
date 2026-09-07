package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/handlers"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/service"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/service/realtime"
	postgres_storage "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/storage/postgres"
	redis_storage "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/storage/redis"
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

	bandClient := clients.NewBandClient()
	centrifugoClient := clients.NewCentrifugoClient()
	employeeInfoClient := clients.NewEmployeeInfoApiClient()
	chatStorage := postgres_storage.NewChatStorage()
	cache := redis_storage.NewCache()
	activityStorage := redis_storage.NewActivityStorage(cache)
	accessStorage := redis_storage.NewAccessStorage(cache)

	runner.RegisterConfigurableEntity(bandClient.Configure)
	runner.RegisterConfigurableEntity(centrifugoClient.Configure)
	runner.RegisterConfigurableEntity(employeeInfoClient.Configure)
	runner.RegisterConfigurableEntity(cache.Configure)

	employeeInfo := service.NewEmployeeInfo(employeeInfoClient)

	srvChat := service.New(bandClient, employeeInfo)
	chatRealtime := realtime.New(bandClient, centrifugoClient, chatStorage, activityStorage, accessStorage)

	chatHandlers := handlers.New(srvChat)

	runner.RegisterServerStartTasks(chatRealtime.Start)

	runner.RegisterCustomHandler("CreateChat", chatHandlers.CreateChat)
	runner.RegisterCustomHandler("GetChat", chatHandlers.GetChat)
	runner.RegisterCustomHandler("GetHistoryChat", chatHandlers.GetHistoryChat)
	runner.RegisterCustomHandler("CreatePost", chatHandlers.CreatePost)
	runner.RegisterCustomHandler("UploadFile", chatHandlers.UploadFile)
	runner.RegisterCustomHandler("DownloadFile", chatHandlers.DownloadFile)

	runner.Run(loggerToClickhouse...)
}
