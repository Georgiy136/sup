package closewh

import (
	closewhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh/models"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
)

func prepareRequestForCloseWh(ext closewhmodels.ExtTicketInfoForCloseWh) closewhmodelsnew.RequestForCloseWh {
	return closewhmodelsnew.RequestForCloseWh{
		OfficeId: ext.OfficeId.Id,
		WhId:     ext.WhId.Id,
	}
}

func prepareRequestForGetStoragePlacesByWh(ext closewhmodels.ExtTicketInfoForCheckUndeletedStoragePlaces) closewhmodelsnew.RequestGetStoragePlacesByWh {
	return closewhmodelsnew.RequestGetStoragePlacesByWh{
		OfficeId: ext.OfficeId.Id,
		WhId:     ext.WhId.Id,
	}
}
