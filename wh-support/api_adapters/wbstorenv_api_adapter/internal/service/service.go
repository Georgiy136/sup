package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	support_utils "gitlab.wildberries.ru/wbwh/support/utils.git/localization"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
)

type ProxyService struct {
	client             *clients.WbstorenvClient
	localizationClient *clients.LocalizationApiClient
	errBuilder         core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (p *ProxyService) Configure(ctx context.Context, config configs.Config) {
	p.client = clients.NewWbstorenvClient(ctx, config)
}

func (p *ProxyService) GetWarehouseInfoByWhID(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		WhID int64 `json:"wh_id" binding:"required,WhID"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	warehouseInfo, err := p.client.GetWarehouseInfoByWhID(ctx, body.WhID)
	if err != nil {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get info about warehouse: %w", err))
		return
	}

	if warehouseInfo == nil {
		utils.BindObjectToRestData(ctx, models.GetWarehouseInfoByWhIDResponse{
			IsValid: false,
			ErrMsg:  p.getLocalizedMessage(ctx, common.WhNotFound),
		})
		return
	}

	if warehouseInfo.IsNotActive == true {
		utils.BindObjectToRestData(ctx, models.GetWarehouseInfoByWhIDResponse{
			IsValid: false,
			ErrMsg:  p.getLocalizedMessage(ctx, common.WhAlreadyInactive),
		})
		return
	}

	if warehouseInfo.IsExistVirtualWh == true {
		utils.BindObjectToRestData(ctx, models.GetWarehouseInfoByWhIDResponse{
			IsValid: false,
			ErrMsg:  p.getLocalizedMessage(ctx, common.WhHasActiveVirtual),
		})
		return
	}

	utils.BindObjectToRestData(ctx, models.GetWarehouseInfoByWhIDResponse{
		IsValid:       true,
		WarehouseInfo: []models.WarehouseInfo{*warehouseInfo},
	})
}

func (p *ProxyService) getLocalizedMessage(ctx *gin.Context, dataError rest_data.CustomError) string {
	lang := support_utils.GetLanguage(ctx)

	if lang == support_utils.DefaultLang {
		return dataError.Message
	}
	keysWithValues := models.KeysWithValues{
		KeyId:         dataError.ErrorKey,
		MessageValues: dataError.MessageValues,
	}

	translatedMsg, err := p.localizationClient.GetTranslation(ctx, lang, keysWithValues)
	if err != nil {
		if errors.Is(err, common.ErrTranslationNotFound) {
			logrus.Errorf("translation not found for key %s: %v", dataError.ErrorKey, err)
			return dataError.Message
		}
		logrus.Errorf("get localized message error: %v", err)
		return dataError.Message
	}

	return translatedMsg
}
