package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_warehouse_info_api_adapter/internal/handlers"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
)

func main() {
	service := handlers.New()
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.RegisterConfigurableEntity(service.Configure)

	runner.RegisterCustomHandler("ListWarehousesOffices", service.ListWarehousesOffices)
	runner.RegisterCustomHandler("GetSelectorOffices", service.GetSelectorOffices)
	runner.RegisterCustomHandler("GetSelectorOfficesFilteredByCountryCode", service.GetSelectorOfficesFilteredByCountryCode)
	runner.RegisterCustomHandler("GetSelectorWh", service.GetSelectorWh)
	runner.RegisterCustomHandler("GetSelectorWhWithoutDel", service.GetSelectorWhWithoutDel)
	runner.RegisterCustomHandler("CheckWhAbbreviation", service.CheckWhAbbreviation)
	runner.RegisterCustomHandler("CheckWhName", service.CheckWhName)
	runner.RegisterCustomHandler("BusinessProcessesGetAll", service.BusinessProcessesGetAll)
	runner.RegisterCustomHandler("ShkStateGetAll", service.ShkStateGetAll)
	runner.RegisterCustomHandler("ShkStateCheckByID", service.ShkStateCheckByID)
	runner.RegisterCustomHandler("GetSelectorBranchOffices", service.GetSelectorBranchOffices)
	runner.Run()
}
