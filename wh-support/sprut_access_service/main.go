package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/clients"
	sprut_client "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/clients/sprut"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/service"
	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	runner := rest_auto_api.NewService()

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	backendApiClient := sprut_client.NewWhBackendApiClient()
	sprutClient := sprut_client.NewSprutClient()
	sprutAccessSrv := sprut_client.NewSprutAccessService(backendApiClient, sprutClient)

	localizationClient := clients.NewLocalizationApiClient()

	apiService := service.NewApiService(sprutAccessSrv, localizationClient)

	runner.RegisterConfigurableEntity(apiService.Configure)

	runner.RegisterCustomHandler("GetBuildingsSelectorByOfficeID", apiService.GetBuildingsSelectorByOfficeID)
	runner.RegisterCustomHandler("CheckBuildingNotExists", apiService.CheckBuildingNotExists)
	runner.RegisterCustomHandler("GetOfficeStagePartByWhID", apiService.GetOfficeStagePartByWhID)
	runner.RegisterCustomHandler("GetStoragePlacesBySections", apiService.GetStoragePlacesBySections)
	runner.RegisterCustomHandler("GetPartsByStage", apiService.GetPartsByStage)
	runner.RegisterCustomHandler("GetStreetsByStage", apiService.GetStreetsByStage)

	runner.Run(loggerToClickhouse...)
}
