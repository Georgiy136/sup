package createoffice

import createofficemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_office/models"

func convertExtCreateOfficeRequestForCreateOffice(ext createofficemodels.ExtTicketInfoForCreateOffice) createofficemodels.RequestForCreateOffice {
	return createofficemodels.RequestForCreateOffice{
		SprutName: ext.SprutName,
		OfficeId:  ext.OfficeId.Id,
	}
}
