package internalresort

import (
	"errors"
	"unicode/utf8"

	internalresortmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/internal_resort/models"
)

const (
	maxCommentSizeSymbol = 750
)

func getPerformValidGoodsIDs(validGoodsIDs map[int64]struct{}) []internalresortmodels.ValidGoodsID {
	const (
		validGoodsIDOrderID       = 1
		validGoodsIDFrontDataName = "ШК для пересорта"
	)

	performValidGoodsIDs := make([]internalresortmodels.ValidGoodsID, 0, len(validGoodsIDs))
	for goodsID := range validGoodsIDs {
		performGoodsID := internalresortmodels.GoodsID{
			Value:         goodsID,
			OrderID:       validGoodsIDOrderID,
			FrontDataName: validGoodsIDFrontDataName,
		}

		performValidGoodsIDs = append(performValidGoodsIDs, internalresortmodels.ValidGoodsID{GoodsID: performGoodsID})
	}
	return performValidGoodsIDs
}

func getPerformNotValidGoodsIDs(notValidGoodsIDs map[string][]int64) []internalresortmodels.NotValidGoodsID {
	const (
		performCommentOrderID       = 1
		performCommentFrontDataName = "Комментарий по невалидному ШК"

		performGoodIDsOrderID       = 2
		performGoodIDsFrontDataName = "ID невалидных ШК"
	)

	performNotValidGoodsIDs := make([]internalresortmodels.NotValidGoodsID, 0, len(notValidGoodsIDs))
	for comment, goodsIDs := range notValidGoodsIDs {
		performComment := internalresortmodels.Comment{
			Value:         comment,
			OrderID:       performCommentOrderID,
			FrontDataName: performCommentFrontDataName,
		}

		performGoodIDs := internalresortmodels.GoodsIDs{
			Value:         goodsIDs,
			OrderID:       performGoodIDsOrderID,
			FrontDataName: performGoodIDsFrontDataName,
		}

		performNotValidGoodsIDs = append(performNotValidGoodsIDs, internalresortmodels.NotValidGoodsID{Comment: performComment, GoodsIDs: performGoodIDs})
	}

	return performNotValidGoodsIDs
}

func validateCommentSize(comment string) error {
	if utf8.RuneCountInString(comment) > maxCommentSizeSymbol {
		return errors.New("invalid comment size")
	}

	return nil
}
