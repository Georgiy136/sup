package handlers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_warehouse_info_api_adapter/internal/common"
	support_utils "gitlab.wildberries.ru/wbwh/support/utils.git/localization"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"

	"errors"
	"strconv"

	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_warehouse_info_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_warehouse_info_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
)

type ProxyService struct {
	client              *clients.WbstorenvWarehouseInfoApiClient
	localizationClient  *clients.LocalizationApiClient
	errBuilder          core_errors.ErrorBuilder
	allowedCountryCodes map[string]struct{}
	validator           *validator.Validate
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder:          core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
		allowedCountryCodes: make(map[string]struct{}),
		validator:           validator.New(),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewWbstorenvWarehouseInfoApiClient(ctx, config)
	s.localizationClient = clients.NewLocalizationApiClient(ctx, config)

	var conf models.ParamsKeys
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("service_param_keys"), &conf); err != nil {
		logrus.Panicf("error parsing service_param_keys configs - %v", err)
	}

	if err := s.validator.Struct(conf); err != nil {
		logrus.Panicf("error validate service_param_keys configs - %v", err)
	}

	for _, val := range conf.AllowedCountryCodes {
		s.allowedCountryCodes[val] = struct{}{}
	}

	logrus.Infof("allowed country codes for api wh/get_selector (GetSelectorOffices): %v", conf.AllowedCountryCodes)
}

func (s *ProxyService) ListWarehousesOffices(ctx *gin.Context, _ map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)
	resp, err := s.client.ListWarehousesOffices(ctx, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get list warehouses offices: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp.Data)
}

func (s *ProxyService) GetSelectorOffices(ctx *gin.Context, _ map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)
	resp, err := s.client.ListWarehousesOffices(ctx, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about offices: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorOffices := make([]models.SelectorOffices, 0, len(resp.Data))
	for i := range resp.Data {
		selectorOffices = append(selectorOffices, models.SelectorOffices{
			OfficeID:   resp.Data[i].OfficeID,
			OfficeName: resp.Data[i].OfficeName,
		})
	}

	utils.BindObjectToRestData(ctx, selectorOffices)
}

func (s *ProxyService) GetSelectorOfficesFilteredByCountryCode(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	resp, err := s.client.ListWarehousesOffices(ctx, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about offices: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorOffices := make([]models.SelectorOffices, 0, len(resp.Data))
	for i := range resp.Data {
		if _, ok := s.allowedCountryCodes[resp.Data[i].CountryCode]; !ok {
			continue
		}

		selectorOffices = append(selectorOffices, models.SelectorOffices{
			OfficeID:   resp.Data[i].OfficeID,
			OfficeName: resp.Data[i].OfficeName,
		})
	}

	utils.BindObjectToRestData(ctx, selectorOffices)
}

func (s *ProxyService) GetSelectorWh(ctx *gin.Context, params map[string]interface{}) {
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

	officeIDParsed := strconv.FormatInt(body.OfficeID, 10)

	resp, err := s.client.GetWHByOffice(ctx, officeIDParsed)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about wh: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorWH := make([]models.SelectorWH, 0, len(resp.Data))
	for i := range resp.Data {
		selectorWH = append(selectorWH, models.SelectorWH{
			WhID:   resp.Data[i].WhID,
			WhName: resp.Data[i].WhName,
		})
	}

	utils.BindObjectToRestData(ctx, selectorWH)
}

func (s *ProxyService) GetSelectorWhWithoutDel(ctx *gin.Context, params map[string]interface{}) {
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

	resp, err := s.client.GetWHByOfficeV2(ctx, body.OfficeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get wh by office V2: %w", err))
		return
	}

	if len(resp) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorWH := make([]models.SelectorWH, 0, len(resp))
	for i := range resp {
		selectorWH = append(selectorWH, models.SelectorWH{
			WhID:   resp[i].WhID,
			WhName: resp[i].WhName,
		})
	}

	utils.BindObjectToRestData(ctx, selectorWH)
}

func (s *ProxyService) CheckWhAbbreviation(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		PlaceNamePrefix string `json:"value" binding:"required,gt=0,lte=5"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	isValidWhAbbreviation, msgErr, err := s.client.CheckNewWhAbbreviation(ctx, body.PlaceNamePrefix)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check new wh abbreviation: %w", err))
		return
	}

	response := models.ResponseWithMsgError{
		IsValid: isValidWhAbbreviation,
		ErrMsg:  msgErr,
	}

	utils.BindObjectToRestData(ctx, response)
}

func (s *ProxyService) CheckWhName(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		WhName string `json:"value" binding:"required,gt=0,lte=128"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	isValidWhName, msgErr, err := s.client.CheckNewWhName(ctx, body.WhName)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check new wh name: %w", err))
		return
	}

	response := models.ResponseWithMsgError{
		IsValid: isValidWhName,
		ErrMsg:  msgErr,
	}

	utils.BindObjectToRestData(ctx, response)
}

func (s *ProxyService) BusinessProcessesGetAll(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.BusinessProcessesGetAll(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about business processes: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp.Data)
}

func (s *ProxyService) ShkStateGetAll(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.ShkStateGetAll(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about shk states: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp.Data)
}

func (s *ProxyService) ShkStateCheckByID(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		StateID string `json:"value" binding:"required,len=3"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	isValidShkState, msgErr, err := s.client.ShkStateCheckByID(ctx, body.StateID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("shk state check by ID error: %w", err))
		return
	}

	response := models.ShkStateStatus{
		IsValid: isValidShkState,
		ErrMsg:  msgErr,
	}

	utils.BindObjectToRestData(ctx, response)
}

func (s *ProxyService) GetSelectorBranchOffices(ctx *gin.Context, params map[string]interface{}) {
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

	branchOffices, err := s.client.GetBranchOfficeByOfficeIDs(ctx, []int64{body.OfficeID}, false)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get branch offices by office ids: %w", err))
		return
	}

	if branchOffices == nil || len(branchOffices.Data) == 0 {
		utils.BindObjectToRestData(ctx, models.ResponseGetSelectorBranchOffices{
			IsValid: false,
			ErrMsg:  s.getLocalizedMessage(ctx, common.OfficeNotFound),
		})
		return
	}

	utils.BindObjectToRestData(ctx, models.ResponseGetSelectorBranchOffices{
		IsValid:       true,
		BranchOffices: branchOffices.Data,
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
		if errors.Is(err, clients.ErrTranslationNotFound) {
			logrus.Errorf("translation not found for key %s: %v", dataError.ErrorKey, err)
			return dataError.Message
		}
		logrus.Errorf("get localized message error: %v", err)
		return dataError.Message
	}

	return translatedMsg
}
