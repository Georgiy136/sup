package closewh

import (
	closewhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
)

const defaultScenarioOrderID = 0

const maxPlacesRemains = 5000

var returnFromIncidentAnalysisStatuses = map[string]struct{}{
	"IN2": {},
	"IN3": {},
	"IN4": {},
	"IN6": {},
	"IN9": {},
	"I11": {},
}

const (
	scenarioGoodsSaveGoodsExcel = 1
	scenarioGoodsLimitExceeded  = 2
	scenarioGoodsLowTotalPrice  = 3
)

const (
	scenarioPlacesDeleted      = 1
	scenarioPlacesNotDeleted   = 2
	scenarioReturnFromIncident = 3
	scenarioAllPlacesDeleted   = 4
)

const (
	scenarioBlockClosed      = 1
	scenarioBlockCloseFailed = 2
)

type storagePlaceScenarioKey struct {
	statusID string
	action   closewhmodels.ActionCode
	reason   string
}

var storagePlaceScenarios = map[storagePlaceScenarioKey]int64{
	{
		statusID: statusIdDeleteStoragePlaceDLP,
		action:   closewhmodels.ActionCompleted,
	}: scenarioPlacesDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL1,
		action:   closewhmodels.ActionCompleted,
	}: scenarioPlacesDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL2,
		action:   closewhmodels.ActionCompleted,
	}: scenarioPlacesDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL3,
		action:   closewhmodels.ActionCompleted,
	}: scenarioPlacesDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL4,
		action:   closewhmodels.ActionCompleted,
	}: scenarioPlacesDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL1,
		action:   closewhmodels.ActionCompletedWithErrors,
		reason:   closewhmodels.WithErrReasonStoragePlacesNotDeleted,
	}: scenarioPlacesNotDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL2,
		action:   closewhmodels.ActionCompletedWithErrors,
		reason:   closewhmodels.WithErrReasonStoragePlacesNotDeleted,
	}: scenarioPlacesNotDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL3,
		action:   closewhmodels.ActionCompletedWithErrors,
		reason:   closewhmodels.WithErrReasonStoragePlacesNotDeleted,
	}: scenarioPlacesNotDeleted,
	{
		statusID: statusIdDeleteStoragePlacePL4,
		action:   closewhmodels.ActionCompletedWithErrors,
		reason:   closewhmodels.WithErrReasonStoragePlacesNotDeleted,
	}: scenarioPlacesNotDeleted,
}

func resolveGoodsScenario(action closewhmodels.ActionCode, reason string, goodsCount int64, totalPriceSum float64) int64 {
	switch {
	case action == closewhmodels.ActionFailed && reason == closewhmodels.FailReasonLimitExceeded:
		return scenarioGoodsLimitExceeded
	case action == closewhmodels.ActionCompleted:
		if goodsCount > goodsMaxTotal {
			return scenarioGoodsLimitExceeded
		}
		if totalPriceSum <= goodsMaxTotalPriceSum {
			return scenarioGoodsLowTotalPrice
		}
		return scenarioGoodsSaveGoodsExcel
	}
	return defaultScenarioOrderID
}

func isDeleteStoragePlaceRetryStatus(statusID string) bool {
	switch statusID {
	case statusIdDeleteStoragePlacePL1, statusIdDeleteStoragePlacePL2,
		statusIdDeleteStoragePlacePL3, statusIdDeleteStoragePlacePL4:
		return true
	default:
		return false
	}
}

func resolveStoragePlaceScenario(statusID string, action closewhmodels.ActionCode, reason string) int64 {
	scenarioOrderID, found := storagePlaceScenarios[storagePlaceScenarioKey{
		statusID: statusID,
		action:   action,
		reason:   reason,
	}]
	if !found {
		return defaultScenarioOrderID
	}

	return scenarioOrderID
}

func resolveCloseWhScenario(action closewhmodels.ActionCode) int64 {
	if action == closewhmodels.ActionFailed {
		return scenarioBlockCloseFailed
	}
	return scenarioBlockClosed
}

func resolveUndeletedStoragePlacesScenario(notDeletedStoragePlace []closewhmodels.NotDeletedStoragePlace, returnFromStatusID string) int64 {
	for _, place := range notDeletedStoragePlace {
		if place.Reason == closewhmodels.ReasonStockRemains && place.PlacesQty > maxPlacesRemains {
			return scenarioPlacesNotDeleted
		}
	}
	if _, ok := returnFromIncidentAnalysisStatuses[returnFromStatusID]; ok {
		return scenarioReturnFromIncident
	}
	return scenarioAllPlacesDeleted
}
