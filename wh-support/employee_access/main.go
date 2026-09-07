package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_access/internal/service"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	runner := rest_auto_api.NewService()

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	srv := service.NewEmployeeAccessService()
	runner.RegisterConfigurableEntity(srv.Init)

	runner.RegisterCustomHandler("GetEmployeeAllowedActionsHandler", srv.GetEmployeeAllowedActionsHandler)
	runner.Run()
}
