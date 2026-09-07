package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/finance_tariff_cluster_api_adapter/internal/handlers"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	service := handlers.New()
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.RegisterConfigurableEntity(service.Configure)

	runner.RegisterCustomHandler("GetAllProdTypes", service.GetAllProdTypesHandler)
	runner.RegisterCustomHandler("GetSelectorProdTypeParts", service.GetSelectorProdTypePartsHandler)

	runner.Run()
}
