package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/client"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/controller"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/service"
	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	runner := rest_auto_api.NewService()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	keycloakClient := client.NewKeycloak()
	keycloakService := service.NewKeycloak(keycloakClient)
	keycloakController := controller.NewKeycloak(keycloakService)

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.RegisterConfigurableEntity(keycloakClient.Configure)
	runner.RegisterConfigurableEntity(keycloakService.Configure)

	runner.RegisterCustomHandler("AuthKeycloakByPassword", keycloakController.AuthKeycloakByPassword)

	runner.Run(loggerToClickhouse...)
}
