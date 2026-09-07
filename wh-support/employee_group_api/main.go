package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/service"
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

	client := clients.NewEmployeeInfoApiClient()
	srv := service.NewPersonalAccount(client)

	runner.RegisterConfigurableEntity(srv.Configure)

	runner.RegisterCustomHandler("GetEmployeesByGroup", srv.GetEmployeesByGroup)
	runner.RegisterCustomHandler("GetGroupsByEmployee", srv.GetGroupsByEmployee)
	runner.RegisterCustomHandler("CreateGroup", srv.CreateGroup)
	runner.RegisterCustomHandler("DeleteGroup", srv.DeleteGroup)
	runner.RegisterCustomHandler("GetAllGroups", srv.GetAllGroups)
	runner.RegisterCustomHandler("AddEmployeeToGroup", srv.AddEmployeeToGroup)
	runner.RegisterCustomHandler("DeleteEmployeeFromGroup", srv.DeleteEmployeeFromGroup)
	runner.RegisterCustomHandler("UpdateExternalActionGroup", srv.UpdateExternalActionGroup)
	runner.RegisterCustomHandler("GetCategoriesWithGroups", srv.GetCategoriesWithGroups)
	runner.RegisterCustomHandler("AddGroupToCategory", srv.AddGroupToCategory)
	runner.RegisterCustomHandler("UpdateGroupForCategory", srv.UpdateGroupForCategory)
	runner.RegisterCustomHandler("DeleteGroupToCategory", srv.DeleteGroupToCategory)
	runner.RegisterCustomHandler("UpdateGroupName", srv.UpdateGroupName)

	runner.Run(loggerToClickhouse...)
}
