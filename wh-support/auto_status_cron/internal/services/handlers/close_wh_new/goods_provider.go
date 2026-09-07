package closewh

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
)

const (
	priceBatchSize  = 1000
	defaultAvgPrice = 1000.00
)

type goodsManager struct {
	storagePlaceProvider storagePlaceProvider
	priceAggregator      priceAggregator
}

func NewGoodsManager(storagePlaceProvider storagePlaceProvider, priceAggregator priceAggregator) *goodsManager {
	return &goodsManager{
		storagePlaceProvider: storagePlaceProvider,
		priceAggregator:      priceAggregator,
	}
}

// FetchAllGoods получает все остатки постранично
func (m *goodsManager) FetchAllGoods(ctx context.Context, whId, officeId int64) ([]closewhmodelsnew.GoodsItem, int64, error) {
	allGoods := make([]closewhmodelsnew.GoodsItem, 0, goodsMaxTotal)
	lastGoodsID := int64(0)

	i := 0
	for {

		if len(allGoods) > goodsMaxTotal {
			break
		}

		pageResp, err := m.storagePlaceProvider.StoragePlaceGetGoodsListByWh(ctx, closewhmodelsnew.RequestGetGoodsPage{
			WhId: whId, OfficeId: officeId, LastGoodsId: lastGoodsID,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("fetch goods page err: %w", err)
		}
		if pageResp == nil {
			return nil, 0, fmt.Errorf("fetch goods page err: resp is nil")
		}

		if len(pageResp.Data) == 0 {
			break
		}

		allGoods = append(allGoods, pageResp.Data...)

		lastGoodsID = pageResp.Data[len(pageResp.Data)-1].GoodsId

		logrus.Infof("wh_id=%d ,office_id=%d, fetch goods page data, iteration=%d, got len(items)=%d", whId, officeId, i, len(pageResp.Data))
		i++
	}

	return allGoods, lastGoodsID, nil
}

// AnnotateGoodsWithPrices обогащает товары ценами
func (m *goodsManager) AnnotateGoodsWithPrices(ctx context.Context, goods []closewhmodelsnew.GoodsItem) ([]closewhmodelsnew.GoodsItemWithPrice, float64, error) {
	if len(goods) == 0 {
		return nil, 0, nil
	}

	nmIDSet := make(map[int64]struct{}, len(goods))
	for _, g := range goods {
		if g.NmId != nil {
			nmIDSet[*g.NmId] = struct{}{}
		}
	}

	nmPriceMap := make(map[int64]float64, len(nmIDSet))
	nmIDs := make([]int64, 0, len(nmIDSet))
	for nmID := range nmIDSet {
		nmIDs = append(nmIDs, nmID)
	}
	for i := 0; i < len(nmIDs); i += priceBatchSize {
		end := i + priceBatchSize
		end = min(end, len(nmIDs))

		priceResp, err := m.priceAggregator.GetAveragePriceNm(ctx, closewhmodelsnew.RequestPriceAggregation{
			NmIds: nmIDs[i:end],
		})
		if err != nil {
			return nil, 0, fmt.Errorf("price aggregation batch [%d:%d] err: %w", i, end, err)
		}
		if priceResp == nil {
			return nil, 0, fmt.Errorf("price aggregation batch [%d:%d] err: empty response", i, end)
		}

		for _, p := range priceResp.Data {
			nmPriceMap[p.NmId] = p.AvgPrice
		}
	}

	result := make([]closewhmodelsnew.GoodsItemWithPrice, 0, len(goods))
	var totalPriceSum float64

	for _, g := range goods {
		avgPrice := defaultAvgPrice
		if g.NmId != nil {
			if p, ok := nmPriceMap[*g.NmId]; ok {
				avgPrice = p
			}
		}
		result = append(result, closewhmodelsnew.GoodsItemWithPrice{
			GoodsId:    g.GoodsId,
			PlaceId:    g.PlaceId,
			SkuId:      g.SkuId,
			NmId:       g.NmId,
			AvgNmPrice: avgPrice,
		})
		totalPriceSum += avgPrice
	}

	return result, totalPriceSum, nil
}

func (m *goodsManager) GoodsExportTaskCreate(ctx context.Context, body closewhmodelsnew.RequestCreateGoodsTask) (closewhmodelsnew.TicketActionResult, error) {
	resp, err := m.storagePlaceProvider.StoragePlaceGoodsExportTaskCreate(ctx, body)
	if err != nil {
		return closewhmodelsnew.TicketActionResult{
			Action: closewhmodelsnew.ActionFailed,
			Reason: closewhmodelsnew.FailReasonAPIError,
		}, err
	}
	switch resp.Data.Status {
	case closewhmodelsnew.GoodsTaskStatusPND, closewhmodelsnew.GoodsTaskStatusPRG:
		return closewhmodelsnew.TicketActionResult{Action: closewhmodelsnew.ActionInProgress}, nil

	case closewhmodelsnew.GoodsTaskStatusERR:
		return closewhmodelsnew.TicketActionResult{
			Action: closewhmodelsnew.ActionFailed,
			Reason: closewhmodelsnew.FailReasonLimitExceeded,
		}, nil

	case closewhmodelsnew.GoodsTaskStatusCMP:
		return closewhmodelsnew.TicketActionResult{Action: closewhmodelsnew.ActionCompleted}, nil
	}
	return closewhmodelsnew.TicketActionResult{
		Action: closewhmodelsnew.ActionFailed,
		Reason: closewhmodelsnew.FailReasonTaskError,
	}, fmt.Errorf("unknown goods task status: %s", resp.Data.Status)
}

func (m *goodsManager) DeleteGoodsFromStoragePlaces(ctx context.Context, body closewhmodelsnew.RequestDeleteStoragePlace) (closewhmodelsnew.TicketActionResult, *closewhmodelsnew.DeleteStoragePlaceData, error) {
	resp, err := m.storagePlaceProvider.DeleteStoragePlacesFromInactiveWh(ctx, body)
	if err != nil {
		return closewhmodelsnew.TicketActionResult{
			Action: closewhmodelsnew.ActionFailed,
			Reason: closewhmodelsnew.FailReasonAPIError,
		}, resp, err
	}
	switch resp.Status {
	case closewhmodelsnew.DeleteStoragePlaceStatusPND, closewhmodelsnew.DeleteStoragePlaceStatusPRG:
		return closewhmodelsnew.TicketActionResult{Action: closewhmodelsnew.ActionInProgress}, resp, nil

	case closewhmodelsnew.DeleteStoragePlaceStatusDEL:
		return closewhmodelsnew.TicketActionResult{Action: closewhmodelsnew.ActionCompleted}, resp, nil

	case closewhmodelsnew.DeleteStoragePlaceStatusERR:
		return closewhmodelsnew.TicketActionResult{
			Action: closewhmodelsnew.ActionCompletedWithErrors,
			Reason: closewhmodelsnew.WithErrReasonStoragePlacesNotDeleted,
		}, resp, nil
	}
	return closewhmodelsnew.TicketActionResult{
		Action: closewhmodelsnew.ActionFailed,
		Reason: closewhmodelsnew.FailReasonAPIError,
	}, resp, fmt.Errorf("unknown delete storage place status: %s", resp.Status)
}

func (m *goodsManager) DeleteGoodsByWh(ctx context.Context, body closewhmodelsnew.RequestDeleteGoodsData) error {
	return m.storagePlaceProvider.StoragePlaceDeleteGoodsByWh(ctx, body)
}

type priceAggregator interface {
	GetAveragePriceNm(ctx context.Context, body closewhmodelsnew.RequestPriceAggregation) (*closewhmodelsnew.ResponsePriceAggregation, error)
}

type storagePlaceProvider interface {
	StoragePlaceGetGoodsListByWh(ctx context.Context, body closewhmodelsnew.RequestGetGoodsPage) (*closewhmodelsnew.ResponseGoodsPage, error)
	DeleteStoragePlacesFromInactiveWh(ctx context.Context, body closewhmodelsnew.RequestDeleteStoragePlace) (*closewhmodelsnew.DeleteStoragePlaceData, error)
	StoragePlaceGoodsExportTaskCreate(ctx context.Context, body closewhmodelsnew.RequestCreateGoodsTask) (*closewhmodelsnew.ResponseGoodsTask, error)
	StoragePlaceDeleteGoodsByWh(ctx context.Context, body closewhmodelsnew.RequestDeleteGoodsData) error
}
