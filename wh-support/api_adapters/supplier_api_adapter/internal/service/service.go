package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/supplier_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/supplier_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/supplier_api_adapter/internal/models"
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
	client             *clients.SupplierApiClient
	localizationClient *clients.LocalizationApiClient
	errBuilder         core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (p *ProxyService) Configure(ctx context.Context, config configs.Config) {
	p.client = clients.NewSupplierApiClient(ctx, config)
}

func (p *ProxyService) GetSupplierInfoByNmIDs(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		NmID int64 `json:"nm_id" binding:"required,gt=0"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	supplierInfo, err := p.client.GetSupplierInfoByNmIDs(ctx, employeeID, []int64{body.NmID})
	if err != nil {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about prodtypes: %w", err))
		return
	}

	if len(supplierInfo) == 0 {
		utils.BindObjectToRestData(ctx, models.GetSupplierInfoByNmIDsResponse{
			IsValid: false,
			ErrMsg:  p.getLocalizedMessage(ctx, common.SupplierNotFound),
		})
		return
	}

	utils.BindObjectToRestData(ctx, models.GetSupplierInfoByNmIDsResponse{
		IsValid:      true,
		SupplierInfo: supplierInfo,
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
