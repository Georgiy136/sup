package closewh

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type TicketHandlerCloseWhNew struct {
	whApi          whApi
	ticketApi      ticketApi
	repo           services.HandlerTicketsRepo
	s3Client       s3Client
	docGen         docGenerator
	fileManagerApi fileManagerApi
	redis          redisClient
	goodsProvider  goodsProvider

	templateDocx []byte
}

func NewTicketHandlerCloseWhNew(
	repo services.HandlerTicketsRepo,
	whApi whApi,
	ticketApi ticketApi,
	s3Client s3Client,
	docGen docGenerator,
	fileManagerApi fileManagerApi,
	redis redisClient,
	goodsProvider goodsProvider,
) *TicketHandlerCloseWhNew {
	return &TicketHandlerCloseWhNew{
		repo:           repo,
		whApi:          whApi,
		ticketApi:      ticketApi,
		s3Client:       s3Client,
		docGen:         docGen,
		fileManagerApi: fileManagerApi,
		redis:          redis,
		goodsProvider:  goodsProvider,
	}
}

func (t *TicketHandlerCloseWhNew) Configure(ctx context.Context, _ configs.Config) {
	var err error
	t.templateDocx, err = t.s3Client.GetObject(ctx, signDocTemplatePath)
	if err != nil {
		logrus.Errorf("can't load template for close wh from s3: %v", err)
	}
	logrus.Infof("template loaded from s3, size: %d bytes", len(t.templateDocx))
}

func (t *TicketHandlerCloseWhNew) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdDeactivateWh:
		if err := t.deactivateWh(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't deactivate wh: %w", err)
		}
	case statusIdDeleteStoragePlaceDLP,
		statusIdDeleteStoragePlacePL1, statusIdDeleteStoragePlacePL2,
		statusIdDeleteStoragePlacePL3, statusIdDeleteStoragePlacePL4:
		if err := t.deleteStoragePlace(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't delete storage place: %w", err)
		}
	case statusIdGetGoodsUP1, statusIdGetGoodsUP2:
		if err := t.getGoods(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't get goods: %w", err)
		}
	case statusIdGenerateSignDocGD1, statusIdGenerateSignDocGD2:
		if err := t.generateSignDoc(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't generate sign doc: %w", err)
		}
	case statusIdCloseWhDl1, statusIdCloseWhDl2, statusIdCloseWhDl3, statusIdCloseWhDl4, statusIdCloseWhDl5:
		if err := t.closeWh(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't close wh: %w", err)
		}
	case statusIdUploadStoragePlacesLD1, statusIdUploadStoragePlacesLD2:
		if err := t.getStoragePlaces(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't get storage places: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}
	return nil
}
