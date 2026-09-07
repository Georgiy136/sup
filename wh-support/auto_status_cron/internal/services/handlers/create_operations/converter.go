package createoperations

import (
	createoperationmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_operations/models"
)

func prepareRequestForProdtypesAddNew(ext createoperationmodels.ExtTicketInfoCreationOperation) createoperationmodels.RequestDataForProdtypesAddNew {
	var req createoperationmodels.RequestDataForProdtypesAddNew

	req.Data = append(req.Data,
		createoperationmodels.RequestForProdtypesAddNew{
			IsCredit:                ext.IsCredit,
			IsDeleted:               false,
			ProdtypeCode:            ext.ProdtypeCode,
			ProdtypeName:            ext.ProdtypeName,
			ProdtypePartID:          ext.ProdtypePartID.ID,
			IsForTarification:       true,
			IsUseImproverMultiplier: ext.IsUseImproverMultiplier,
			RateLimit:               ext.RateLimit,
		},
	)

	return req
}
