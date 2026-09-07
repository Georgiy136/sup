package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/external_mobile_control_pa_adapter/internal/service"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	srv := service.New()

	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	runner.RegisterConfigurableEntity(srv.Configure)
	runner.RegisterCustomHandler("CheckDevice", srv.CheckDevice)

	runner.Run()
}
