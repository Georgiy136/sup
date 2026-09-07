package handlers

import (
	"context"
	"fmt"

	"errors"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_place_info_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_place_info_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
)

const (
	storagePlaceTypeRack               = 1    // Полка
	storagePlaceTypeRackKGT            = 1112 // Полка КБТ
	storagePlaceTypePickingPallet      = 1048 // Палеты Подбор
	storagePlaceTypeSortingPallet      = 1073 // Палеты Подсорт
	storagePlaceTypeOuterwear          = 440  // Верхняя Одежда
	storagePlaceTypeKGT                = 1265 // КГТ
	storagePlaceTypeClothes            = 442  // Вещи
	storagePlaceTypeSGTShelvingPicking = 1701 // Подбор стеллажное хранение СГТ
	storagePlaceTypeFloor1200x1600     = 1702 // Напольное хранение 1200*1600
	storagePlaceTypeFloor1200x2400     = 1703 // Напольное хранение 1200*2400
	storagePlaceTypeFloor1200x3200     = 1704 // Напольное хранение 1200*3200
)

var allowedStoragePlaceTypeIDs = map[int64]struct{}{
	storagePlaceTypeRack:               {},
	storagePlaceTypeRackKGT:            {},
	storagePlaceTypePickingPallet:      {},
	storagePlaceTypeSortingPallet:      {},
	storagePlaceTypeOuterwear:          {},
	storagePlaceTypeKGT:                {},
	storagePlaceTypeClothes:            {},
	storagePlaceTypeSGTShelvingPicking: {},
	storagePlaceTypeFloor1200x1600:     {},
	storagePlaceTypeFloor1200x2400:     {},
	storagePlaceTypeFloor1200x3200:     {},
}

type ProxyService struct {
	client     *clients.WbstorenvPlaceInfoApiClient
	errBuilder core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewWbstorenvPlaceInfoApiClient(ctx, config)
}

func (s *ProxyService) GetStoragePlaceStickerPrefixes(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.GetStoragePlaceStickerPrefixes(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all sticker prefixes: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp.Data)
}

func (s *ProxyService) CheckStoragePlaceType(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		PlaceTypeName string `json:"value" binding:"required,lte=255"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	isValidStoragePlaceType, msgErr, err := s.client.CheckStoragePlaceType(ctx, body.PlaceTypeName)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check place type name: %w", err))
		return
	}

	response := models.ResponseWithMsgError{
		IsValid: isValidStoragePlaceType,
		ErrMsg:  msgErr,
	}

	utils.BindObjectToRestData(ctx, response)
}

func (s *ProxyService) GetNamesOfStoragePlaceTypes(ctx *gin.Context, params map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)
	resp, err := s.client.GetNamesOfStoragePlaceTypes(ctx, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get names of storage place types: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorStoragePlaceTypesName := models.StoragePlaceTypes{
		Data: make([]models.StoragePlaceType, 0, len(resp.Data)),
	}
	for _, v := range resp.Data {
		if _, ok := allowedStoragePlaceTypeIDs[v.PlaceTypeID]; ok {
			selectorStoragePlaceTypesName.Data = append(selectorStoragePlaceTypesName.Data, v)
		}
	}

	if len(selectorStoragePlaceTypesName.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, selectorStoragePlaceTypesName.Data)
}
