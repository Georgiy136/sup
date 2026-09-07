package closewh

import (
	"fmt"
	"strconv"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/docgen"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

func prepareRequestForCloseWh(ext closewhmodelsnew.ExtTicketInfoForCloseWh, employeeID int64) closewhmodelsnew.RequestForCloseWh {
	return closewhmodelsnew.RequestForCloseWh{
		OfficeId:   ext.OfficeId.Id,
		WhId:       ext.WhId.Id,
		EmployeeId: employeeID,
	}
}

func prepareRequestForDeactivateWh(ext closewhmodelsnew.ExtTicketInfoForDeactivateWh, employeeID int64) closewhmodelsnew.RequestForDeactivateWh {
	return closewhmodelsnew.RequestForDeactivateWh{
		OfficeId:   ext.OfficeId.Id,
		WhId:       ext.WhId.Id,
		EmployeeId: employeeID,
	}
}

func prepareRequestForGetStoragePlacesByWh(ext closewhmodelsnew.ExtTicketInfoForGetStoragePlaces) closewhmodelsnew.RequestGetStoragePlacesByWh {
	return closewhmodelsnew.RequestGetStoragePlacesByWh{
		OfficeId: ext.OfficeId.Id,
		WhId:     ext.WhId.Id,
	}
}

func buildPlaceIdsPayload(placeIDs []int64) closewhmodelsnew.PlaceIdsPayload {
	result := make([]closewhmodelsnew.PlaceIdItem, 0, len(placeIDs))
	for i := range placeIDs {
		result = append(result, closewhmodelsnew.PlaceIdItem{
			PlaceId: models.FieldValue{
				Value:         placeIDs[i],
				OrderId:       1,
				FrontDataName: "Идентификатор МХ",
			},
		})
	}
	return closewhmodelsnew.PlaceIdsPayload{PlaceIds: result}
}

func buildNotDeletedPlacesPayload(notDeleted []closewhmodelsnew.NotDeletedStoragePlace) []closewhmodelsnew.StoragePlaceIdsDBSaveData {
	result := make([]closewhmodelsnew.StoragePlaceIdsDBSaveData, 0, len(notDeleted))
	for i := range notDeleted {
		result = append(result, closewhmodelsnew.StoragePlaceIdsDBSaveData{
			PlaceIds: models.FieldValue{OrderId: 1, FrontDataName: "МХ", Value: notDeleted[i].PlaceIds},
			Reason:   models.FieldValue{OrderId: 2, FrontDataName: "Причина", Value: closewhmodelsnew.ReasonToString(notDeleted[i].Reason)},
			Count:    models.FieldValue{OrderId: 3, FrontDataName: "Количество неудаленных МХ", Value: notDeleted[i].PlacesQty},
		})
	}
	return result
}

func buildGoodsWithPriceSheetData(goods []closewhmodelsnew.GoodsItemWithPrice) docgen.ExcelSheetData {
	var (
		goodsIds  = make([]any, 0, len(goods))
		placeIds  = make([]any, 0, len(goods))
		skuIds    = make([]any, 0, len(goods))
		nmIds     = make([]any, 0, len(goods))
		avgPrices = make([]any, 0, len(goods))
	)

	for _, g := range goods {
		goodsIds = append(goodsIds, g.GoodsId)
		placeIds = append(placeIds, g.PlaceId)
		skuIds = append(skuIds, opt(g.SkuId))
		nmIds = append(nmIds, opt(g.NmId))
		avgPrices = append(avgPrices, g.AvgNmPrice)
	}

	return docgen.ExcelSheetData{
		Name: "Остатки по МХ",
		Columns: []docgen.ExcelColumnData{
			{Name: "goods_id", Values: goodsIds},
			{Name: "place_id", Values: placeIds},
			{Name: "sku_id", Values: skuIds},
			{Name: "nm_id", Values: nmIds},
			{Name: "avg_nm_price", Values: avgPrices},
		},
	}
}

func opt[T any](ptr *T) any {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func buildSignDocReplacements(ticketID int64, ext closewhmodelsnew.ExtTicketInfoForGenerateSignDoc) (map[string]string, error) {
	var (
		employeeName         string
		employeeNameGenitive string
	)
	switch ext.OfficeType.Id {
	case officeTypeSC:
		if ext.OfficeType.Name != officeTypeNameSK {
			return nil, fmt.Errorf("office type %d, name not match", ext.OfficeType.Id)
		}
		employeeName = employeeNameSKEmployee
		employeeNameGenitive = employeeNameSKEmployeeGenitive
	case officeTypeSR:
		if ext.OfficeType.Name != officeTypeNameSC {
			return nil, fmt.Errorf("office type %d, name not match", ext.OfficeType.Id)
		}
		employeeName = employeeNameSCEmployee
		employeeNameGenitive = employeeNameSCEmployeeGenitive
	default:
		return nil, fmt.Errorf("unknown office type: %d", ext.OfficeType.Id)
	}

	return map[string]string{
		"$1": time.Now().Format("02.01.2006"),
		"$2": fmt.Sprintf("%.2f", ext.TotalPrice),
		"$3": strconv.FormatInt(ext.TotalCount, 10),
		"$4": strconv.FormatInt(ticketID, 10),
		"$5": strconv.FormatInt(ext.WhId.Id, 10),
		"$6": ext.WhId.Name,
		"$7": employeeName,
		"$8": employeeNameGenitive,
	}, nil
}
