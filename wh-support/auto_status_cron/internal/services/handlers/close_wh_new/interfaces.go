package closewh

import (
	"context"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/docgen"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

type whApi interface {
	CloseWh(ctx context.Context, body closewhmodelsnew.RequestForCloseWh) error
	DeactivateWh(ctx context.Context, body closewhmodelsnew.RequestForDeactivateWh) error
	GetStoragePlacesByWh(ctx context.Context, body closewhmodelsnew.RequestGetStoragePlacesByWh) ([]int64, error)
}

type ticketApi interface {
	PerformTicketV2(ctx context.Context, body models.RequestPerformTicketV2, employeeID int64) error
}

type docGenerator interface {
	RenderTemplateToPdf(ctx context.Context, templateBytes []byte, replacements map[string]string) ([]byte, error)
	GenerateExcel(sheets ...docgen.ExcelSheetData) ([]byte, error)
}

type fileManagerApi interface {
	UploadFile(ctx context.Context, fileBytes []byte, fileName string) (int64, error)
}

type s3Client interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
}

type redisClient interface {
	Exists(ctx context.Context, key string) (bool, error)
	SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

type goodsProvider interface {
	FetchAllGoods(ctx context.Context, whId, officeId int64) ([]closewhmodelsnew.GoodsItem, int64, error)
	AnnotateGoodsWithPrices(ctx context.Context, goods []closewhmodelsnew.GoodsItem) ([]closewhmodelsnew.GoodsItemWithPrice, float64, error)
	DeleteGoodsFromStoragePlaces(ctx context.Context, body closewhmodelsnew.RequestDeleteStoragePlace) (closewhmodelsnew.TicketActionResult, *closewhmodelsnew.DeleteStoragePlaceData, error)
	GoodsExportTaskCreate(ctx context.Context, body closewhmodelsnew.RequestCreateGoodsTask) (closewhmodelsnew.TicketActionResult, error)
	DeleteGoodsByWh(ctx context.Context, body closewhmodelsnew.RequestDeleteGoodsData) error
}
