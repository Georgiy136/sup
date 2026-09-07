package createstreets

import (
	createstreetsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_streets/models"
)

func convertExtInfoToRequestForCreateStreet(ext createstreetsmodels.ExtTicketInfoForCreateStreets, stages []int64, employeeID int64) createstreetsmodels.RequestForCreateStreet {
	return createstreetsmodels.RequestForCreateStreet{
		OfficeID:     ext.OfficeId.Id,
		WhID:         ext.WhId.Id,
		Stages:       stages,
		StreetStart:  ext.StreetStart,
		StreetEnd:    ext.StreetEnd,
		SectionStart: ext.SectionFrom,
		SectionEnd:   ext.SectionEnd,
		EmployeeID:   employeeID,
	}
}
