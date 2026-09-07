package createwh

import (
	createwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_wh/models"
)

func prepareRequestDataForAddNewBuilding(ext createwhmodels.ExtTicketInfoAddBuilding) createwhmodels.RequestDataForAddNewBuilding {
	return createwhmodels.RequestDataForAddNewBuilding{
		IsDel:      false,
		OfficeID:   ext.OfficeId.Id,
		BuildingID: ext.BuildingID,
	}
}

func convertExtTicketInfoPhysicalWhWithCreateBuildingToRequestDataForAddNewWh(ext createwhmodels.ExtTicketInfoAddPhysicalWhWithCreateBuilding) createwhmodels.RequestDataForAddNewWh {
	return createwhmodels.RequestDataForAddNewWh{
		WhName:                   ext.WhName,
		OfficeId:                 ext.OfficeId.Id,
		TimeZone:                 ext.TimeZone.Id,
		BuildingId:               &ext.BuildingId,
		PlaceNamePrefix:          ext.PlaceNamePrefix,
		IsAssemblyByDeliveryDate: ext.IsAssemblyByDeliveryDate,
	}
}

func convertExtTicketInfoPhysicalWhWithSelectBuildingToRequestDataForAddNewWh(ext createwhmodels.ExtTicketInfoAddPhysicalWhWithSelectBuilding) createwhmodels.RequestDataForAddNewWh {
	return createwhmodels.RequestDataForAddNewWh{
		WhName:                   ext.WhName,
		OfficeId:                 ext.OfficeId.Id,
		TimeZone:                 ext.TimeZone.Id,
		BuildingId:               &ext.BuildingId.Id,
		PlaceNamePrefix:          ext.PlaceNamePrefix,
		IsAssemblyByDeliveryDate: ext.IsAssemblyByDeliveryDate,
	}
}

func convertExtTicketInfoAddVirtualWhAW3ToRequestDataForAddNewWh(ext createwhmodels.ExtTicketInfoAddVirtualWhAW3) createwhmodels.RequestDataForAddNewWh {
	return createwhmodels.RequestDataForAddNewWh{
		WhName:          ext.WhName,
		OfficeId:        ext.OfficeId.Id,
		TimeZone:        ext.TimeZone.Id,
		PlaceNamePrefix: ext.PlaceNamePrefix,
		PhysicalWhId:    &ext.PhysicalWhId.Id,
		IsWhForTmc:      ext.IsWhForTmc,
	}
}

func convertExtTicketInfoAddVirtualWhAW4ToRequestDataForAddNewWh(ext createwhmodels.ExtTicketInfoAddVirtualWhAW4) createwhmodels.RequestDataForAddNewWh {
	return createwhmodels.RequestDataForAddNewWh{
		WhName:          ext.WhName,
		OfficeId:        ext.OfficeId.Id,
		TimeZone:        ext.TimeZone.Id,
		PlaceNamePrefix: ext.PlaceNamePrefix,
		PhysicalWhId:    &ext.PhysicalWhId.Id,
	}
}
