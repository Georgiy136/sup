package excludetare

import (
	excludetaremodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_tare/models"
)

func prepareRequestDataForExcludeTare(ext excludetaremodels.ExtTicketInfoForExcludeTare) excludetaremodels.RequestBodyForExcludeTare {
	return excludetaremodels.RequestBodyForExcludeTare{
		TareExcel:      convertTareFromDBFormat(ext.TareExcel),
		ExclusionHours: ext.ExclusionHours,
	}
}

func prepareRequestDataForExcludeTareOnWriteoffApi(ext excludetaremodels.ExtTicketInfoForExcludeTare, employeeId int64) []excludetaremodels.RequestBodyForExcludeTareOnWriteoffApi {
	return []excludetaremodels.RequestBodyForExcludeTareOnWriteoffApi{
		{
			TareExcel:      convertTareFromDBFormat(ext.TareExcel),
			ExclusionHours: ext.ExclusionHours,
			EmployeeID:     employeeId,
		},
	}
}

func convertTareFromDBFormat(tareDB []excludetaremodels.TareFromDB) []excludetaremodels.Tare {
	tare := make([]excludetaremodels.Tare, len(tareDB))
	for i := range tare {
		tare[i] = excludetaremodels.Tare{
			TareId:   tareDB[i].TareId.Value,
			TareType: tareDB[i].TareType.Value,
		}
	}
	return tare
}
