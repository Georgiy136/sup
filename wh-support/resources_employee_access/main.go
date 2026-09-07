package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/service"
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
	runner.EnableAuthActionsCheck()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	srv := service.NewResourcesService()

	runner.RegisterConfigurableEntity(srv.Init)

	runner.RegisterCustomHandler("GetResources", srv.GetResourcesHandler)
	runner.RegisterCustomHandler("GetAllActions", srv.GetAllActionsHandler)
	runner.RegisterCustomHandler("CreateAction", srv.CreateActionHandler)
	runner.RegisterCustomHandler("UpdateAction", srv.UpdateActionHandler)
	runner.RegisterCustomHandler("DeleteAction", srv.DeleteActionHandler)
	runner.RegisterCustomHandler("GetEmployeesByActionHandler", srv.GetEmployeesByActionHandler)
	runner.RegisterCustomHandler("GetAccessActionsByEmployeeID", srv.GetAccessActionsByEmployeeIDHandler)
	runner.RegisterCustomHandler("GetTypeActions", srv.GetTypeActionsHandler)
	runner.RegisterCustomHandler("AddTypeAction", srv.AddTypeActionHandler)
	runner.RegisterCustomHandler("DeleteTypeAction", srv.DeleteTypeActionHandler)

	runner.Run(loggerToClickhouse...)
}
