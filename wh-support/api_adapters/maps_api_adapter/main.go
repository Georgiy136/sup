package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/maps_api_adapter/internal/service"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	srv := service.New()

	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	runner.RegisterConfigurableEntity(srv.Configure)
	runner.RegisterCustomHandler("GetSpruts", srv.GetSpruts)
	runner.RegisterCustomHandler("GetOffices", srv.GetOffices)
	runner.RegisterCustomHandler("GetFilteredOffices", srv.GetFilteredOffices)

	runner.Run()
}
