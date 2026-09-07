package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/handlers"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	service := handlers.New()
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.RegisterConfigurableEntity(service.Configure)

	runner.RegisterCustomHandler("GetSprutNamesV001", service.GetSprutNamesHandler)
	runner.RegisterCustomHandler("CheckSprutOfficeExists", service.CheckExistOfficeHandler)
	runner.RegisterCustomHandler("CheckSprutOfficeNotExists", service.CheckNotExistOfficeHandler)
	runner.RegisterCustomHandler("CheckSprutOfficeExistsV2", service.CheckSprutOfficeExistsV2Handler)
	runner.RegisterCustomHandler("CheckSprutOfficeNotExistsV2", service.CheckSprutOfficeNotExistsV2Handler)
	runner.RegisterCustomHandler("GetBranchOfficeByIDWithSprutCheck", service.GetBranchOfficeByIDWithSprutCheck)

	runner.Run()
}
