package transferstreetsbetweenwh

import (
	"fmt"
	"slices"

	transferstreetsbetweenwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/transfer_streets_between_wh/models"
)

func convertExtInfoToRequestForTransferWhForStock(ext transferstreetsbetweenwhmodels.ExtTicketInfoForTransferStreetsBetweenWh, employeeID, streetsSectionsIndex int64) (transferstreetsbetweenwhmodels.RequestForTransferWhForStock, error) {
	if streetsSectionsIndex >= int64(len(ext.StreetsSections)) {
		return transferstreetsbetweenwhmodels.RequestForTransferWhForStock{}, fmt.Errorf("streetsSectionsIndex out of range exception")
	}

	var (
		streetStart = slices.Min(ext.StreetsSections[streetsSectionsIndex].Streets.Value)
		streetEnd   = slices.Max(ext.StreetsSections[streetsSectionsIndex].Streets.Value)

		sectionStart = slices.Min(ext.StreetsSections[streetsSectionsIndex].Sections.Value)
		sectionEnd   = slices.Max(ext.StreetsSections[streetsSectionsIndex].Sections.Value)
	)

	return transferstreetsbetweenwhmodels.RequestForTransferWhForStock{
		OfficeID:     ext.OfficeId.Id,
		OldWhID:      ext.WhIdOld.Id,
		NewWhID:      ext.WhIdNew.Id,
		Stage:        ext.StreetsSections[streetsSectionsIndex].Stage.Value.Id,
		OldPart:      ext.StreetsSections[streetsSectionsIndex].OldPart.Value.Id,
		NewPart:      ext.StreetsSections[streetsSectionsIndex].NewPart.Value.Id,
		StreetStart:  streetStart,
		StreetEnd:    streetEnd,
		SectionStart: sectionStart,
		SectionEnd:   sectionEnd,
		IsReverse:    ext.StreetsSections[streetsSectionsIndex].IsReverse.Value,
		EmployeeID:   employeeID,
	}, nil
}

func convertRequestToErrCommentValue(request transferstreetsbetweenwhmodels.RequestForTransferWhForStock) transferstreetsbetweenwhmodels.ErrCommentValue {
	return transferstreetsbetweenwhmodels.ErrCommentValue{
		Stage:        request.Stage,
		StreetStart:  request.StreetStart,
		StreetEnd:    request.StreetEnd,
		SectionStart: request.SectionStart,
		SectionEnd:   request.SectionEnd,
	}
}
