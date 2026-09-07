package commonhandlers

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"

	"github.com/go-playground/validator/v10"
)

type CommonHandlers struct {
	repo        services.HandlerTicketsRepo
	stageApi    stageApi
	assemblyApi assemblyApi
	inventApi   inventApi

	validator *validator.Validate
}

func NewCommonHandlers(repo services.HandlerTicketsRepo, stage stageApi, assemblyApi assemblyApi, inventApi inventApi) *CommonHandlers {
	return &CommonHandlers{
		repo:        repo,
		stageApi:    stage,
		assemblyApi: assemblyApi,
		inventApi:   inventApi,
		validator:   validator.New(),
	}
}

type stageApi interface {
	CreateStage(ctx context.Context, body commonhandlersmodels.BodyRequestForCreateStage) error
	CreatePart(ctx context.Context, body commonhandlersmodels.BodyRequestForCreatePart) error
}

type assemblyApi interface {
	AssemblyStreetExclusionUpdate(ctx context.Context, body commonhandlersmodels.BodyRequestForExclusionUpdateStreet) error
	StreetExclusionFromSaleUpdateV002(ctx context.Context, body commonhandlersmodels.BodyRequestForStreetExclusionFromSaleUpdateV002) error
	StageOrWhExclusionFromAssemblyUpdateV001(ctx context.Context, body commonhandlersmodels.RequestForStageOrWhExclusionFromAssemblyUpdateV001) error
}

type inventApi interface {
	AddGoodsOnStockInventTask(ctx context.Context, body commonhandlersmodels.BodyRequestForAddGoodsOnStockTask) ([]int64, error)
}
