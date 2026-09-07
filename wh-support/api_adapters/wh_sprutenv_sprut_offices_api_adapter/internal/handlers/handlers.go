package handlers

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/common"
	support_utils "gitlab.wildberries.ru/wbwh/support/utils.git/localization"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

	"errors"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProxyService struct {
	sprutenvSprutOfficesApi       *clients.WhSprutenvSprutOfficesApiClient
	mapsApi                       *clients.MapsApiClient
	storenvWarehouseInfoApiClient *clients.WbstorenvWarehouseInfoApiClient
	localizationClient            *clients.LocalizationApiClient
	errBuilder                    core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.sprutenvSprutOfficesApi = clients.NewWhSprutenvSprutOfficesApiClient(ctx, config)
	s.mapsApi = clients.NewMapsApiClient(ctx, config)
	s.storenvWarehouseInfoApiClient = clients.NewWbstorenvWarehouseInfoApiClient(ctx, config)
	s.localizationClient = clients.NewLocalizationApiClient(ctx, config)
}

func (s *ProxyService) GetSprutNamesHandler(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.sprutenvSprutOfficesApi.GetSprutNames(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get sprut names: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp.Data)
}

func (s *ProxyService) CheckExistOfficeHandler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"value" binding:"required,OfficeID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resp, err := s.sprutenvSprutOfficesApi.CheckExistSprutOffice(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check exist sprut office: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, resp.OfficeExists)
}

func (s *ProxyService) CheckNotExistOfficeHandler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"value" binding:"required,OfficeID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resp, err := s.sprutenvSprutOfficesApi.CheckExistSprutOffice(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check not exist sprut office: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, !resp.OfficeExists)
}

func (s *ProxyService) CheckSprutOfficeExistsV2Handler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"value" binding:"required,OfficeID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	officeOnSprut, err := s.sprutenvSprutOfficesApi.CheckExistSprutOffice(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check exist sprut office: %w", err))
		return
	}

	resp := models.ResponseWithMsgError{
		IsValid: officeOnSprut.OfficeExists,
	}

	if !officeOnSprut.OfficeExists {
		resp.ErrMsg = s.getLocalizedMessage(ctx, common.OfficeNotFound)
	}

	utils.BindObjectToRestData(ctx, resp)
}

func (s *ProxyService) CheckSprutOfficeNotExistsV2Handler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"value" binding:"required,OfficeID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	offices, err := s.mapsApi.GetOfficesWithInactiveByOfficeIDs(ctx, []int64{body.OfficeID})
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check office exists with inactive: %w", err))
		return
	}

	if offices == nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("empty response while check office exists with inactive: %w", err))
		return
	}

	if len(offices.Offices) > 1 {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("too many offices in response maps api: %w", err))
		return
	}

	if len(offices.Offices) == 0 {
		utils.BindObjectToRestData(ctx, models.ResponseWithMsgError{
			IsValid: false,
			ErrMsg:  s.getLocalizedMessage(ctx, common.OfficeNotFound),
		})
		return
	}

	officeOnSprut, err := s.sprutenvSprutOfficesApi.CheckExistSprutOffice(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check not exist sprut office: %w", err))
		return
	}

	resp := models.ResponseWithMsgError{
		IsValid: !officeOnSprut.OfficeExists,
	}

	if officeOnSprut.OfficeExists {
		resp.ErrMsg = s.getLocalizedMessage(ctx, common.OfficeExists)
	}

	utils.BindObjectToRestData(ctx, resp)
}

func (s *ProxyService) GetBranchOfficeByIDWithSprutCheck(ctx *gin.Context, params map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"office_id" binding:"required,OfficeID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	officeInfo, err := s.storenvWarehouseInfoApiClient.GetBranchOfficeInfoByID(ctx, employeeID, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get office info by id: %w", err))
		return
	}
	if officeInfo == nil || len(officeInfo.Offices) == 0 {
		utils.BindObjectToRestData(ctx, models.ResponseGetBranchOffice{
			IsValid: false,
			ErrMsg:  s.getLocalizedMessage(ctx, common.OfficeNotFound),
		})
		return
	}

	var branchOffice []models.BranchOffice
	for _, of := range officeInfo.Offices {
		if of.OfficeID == body.OfficeID {
			branchOffice = append(branchOffice, models.BranchOffice{
				OfficeID:      of.OfficeID,
				OfficeName:    of.OfficeName,
				TypePointName: of.TypePointName,
			})
			break
		}
	}

	if len(branchOffice) == 0 {
		utils.BindObjectToRestData(ctx, models.ResponseGetBranchOffice{
			IsValid: false,
			ErrMsg:  s.getLocalizedMessage(ctx, common.OfficeNotFound),
		})
		return
	}

	officeOnSprut, err := s.sprutenvSprutOfficesApi.CheckExistSprutOffice(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check not exist sprut office: %w", err))
		return
	}

	if officeOnSprut.OfficeExists {
		logrus.Infof("office with id: %d already exist", body.OfficeID)
		utils.BindObjectToRestData(ctx, models.ResponseGetBranchOffice{
			IsValid: false,
			ErrMsg:  s.getLocalizedMessage(ctx, common.OfficeExists),
		})
		return
	}

	utils.BindObjectToRestData(ctx, models.ResponseGetBranchOffice{
		IsValid:       true,
		BranchOffices: branchOffice,
	})
}

func (s *ProxyService) getLocalizedMessage(ctx *gin.Context, dataError rest_data.CustomError) string {
	lang := support_utils.GetLanguage(ctx)

	if lang == support_utils.DefaultLang {
		return dataError.Message
	}
	keysWithValues := models.KeysWithValues{
		KeyId:         dataError.ErrorKey,
		MessageValues: dataError.MessageValues,
	}

	translatedMsg, err := s.localizationClient.GetTranslation(ctx, lang, keysWithValues)
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
