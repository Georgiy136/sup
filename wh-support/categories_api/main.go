package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/service"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/storage/redis"
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

	redisClient := redis.NewRedisClient()
	resourceEmployeeAccessClient := clients.NewResourceEmployeeAccessClient()

	runner.RegisterConfigurableEntity(redisClient.Configure)
	runner.RegisterConfigurableEntity(resourceEmployeeAccessClient.Configure)

	srvCategories := service.NewCategoryInfo(redisClient, resourceEmployeeAccessClient)
	runner.RegisterCustomHandler("GetCategoriesTree", srvCategories.GetCategoriesTree)
	runner.RegisterCustomHandler("GetCategoriesTreeForCreateTicket", srvCategories.GetCategoriesTreeForCreateTicket)

	runner.Run(loggerToClickhouse...)
}
