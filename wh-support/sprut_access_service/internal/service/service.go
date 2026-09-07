package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/clients"
	sprut_client "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/clients/sprut"
	apierrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/models"
	support_utils "gitlab.wildberries.ru/wbwh/support/utils.git/localization"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

type ApiService struct {
	sprutAccessService *sprut_client.SprutAccessService
	localizationClient *clients.LocalizationApiClient
	errBuilder         core_errors.ErrorBuilder
}

func NewApiService(sprutAccessService *sprut_client.SprutAccessService, localizationClient *clients.LocalizationApiClient) *ApiService {
	errorsRepo := errors_keys.NewErrorsRepository()
	errorsRepo.SetErrors(support_err_keys.ErrorsRepository)

	return &ApiService{
		sprutAccessService: sprutAccessService,
		localizationClient: localizationClient,
		errBuilder:         core_errors.NewErrorBuilder(errorsRepo.GetErrors(), nil),
	}

}

func (s *ApiService) Configure(ctx context.Context, cfg configs.Config) {
	s.sprutAccessService.Configure(ctx, cfg)
	s.localizationClient.Configure(ctx, cfg)
}

func (s *ApiService) GetBuildingsSelectorByOfficeID(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

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

	buildingsInfo, err := s.sprutAccessService.GetBuildingsByOfficeID(ctx, body.OfficeID, employeeID)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get buildings info by officeID")
		return
	}

	if len(buildingsInfo) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	buildingsSelector := make([]models.BuildingsSelector, len(buildingsInfo))

	for i := range buildingsInfo {
		buildingsSelector[i].BuildingID = buildingsInfo[i].BuildingID
	}

	utils.BindObjectToRestData(ctx, buildingsSelector)
}

func (s *ApiService) CheckBuildingNotExists(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID   int64  `json:"office_id" binding:"required,OfficeID"`
		BuildingID string `json:"value" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	buildingsInfo, err := s.sprutAccessService.GetBuildingsByOfficeID(ctx, body.OfficeID, employeeID)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get buildings info by officeID")
		return
	}

	resp := models.ResponseWithMsgError{
		IsValid: true,
	}

	for _, building := range buildingsInfo {
		if building.BuildingID == body.BuildingID {
			resp.IsValid = false
			break
		}
	}

	if !resp.IsValid {
		resp.ErrMsg = s.getLocalizedMessage(ctx, apierrors.MsgErrorBuildingExists)
	}

	utils.BindObjectToRestData(ctx, resp)
}

func (s *ApiService) GetOfficeStagePartByWhID(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"office_id" binding:"required,OfficeID"`
		WhID     int64 `json:"wh_id" binding:"required,WhID"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	stagesInfo, err := s.sprutAccessService.GetOfficeStagePartByWhID(ctx, employeeID, body.OfficeID, body.WhID)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get stages info by officeID, whID")
		return
	}

	if len(stagesInfo.Stages) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	stagesSelector := make([]models.StagesSelector, len(stagesInfo.Stages))

	for i := range stagesInfo.Stages {
		stagesSelector[i].Stage = stagesInfo.Stages[i].Stage
	}

	utils.BindObjectToRestData(ctx, stagesSelector)
}

func (s *ApiService) GetStoragePlacesBySections(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID     int64  `json:"office_id" binding:"required,OfficeID"`
		WhID         int64  `json:"wh_id" binding:"required,WhID"`
		Street       int64  `json:"street" binding:"required,gt=0"`
		Stage        *int64 `json:"stage" binding:"omitempty,gte=0"`
		SectionStart int64  `json:"section_start" binding:"required,gt=0"`
		SectionEnd   int64  `json:"section_end" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if body.SectionEnd < body.SectionStart {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(support_err_keys.KeyErrorStreetSectionsInvalid), fmt.Errorf("section_end must be greater than section_start"))
		return
	}

	if body.SectionEnd-body.SectionStart >= 200 {
		s.errBuilder.BindErrorFormat(ctx, core_errors.ErrorKey(support_err_keys.KeyErrorStreetSectionsExceeded), fmt.Errorf("difference between section_end and section_start must be less than 200"), map[string]any{"max_sections": "200"})
		return
	}

	storagePlaces, err := s.sprutAccessService.GetStoragePlacesBySections(ctx, employeeID, body.OfficeID, body.WhID, body.Street, body.Stage, body.SectionStart, body.SectionEnd)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get storage places by sections")
		return
	}

	if len(storagePlaces) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, storagePlaces)
}

func (s *ApiService) GetPartsByStage(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64  `json:"office_id" binding:"required,OfficeID"`
		WhID     int64  `json:"wh_id" binding:"required,WhID"`
		Stage    *int64 `json:"stage" binding:"required,gte=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	stagesInfo, err := s.sprutAccessService.GetOfficeStagePartByWhID(ctx, employeeID, body.OfficeID, body.WhID)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get parts by stage")
		return
	}

	if len(stagesInfo.Stages) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	partsSelector := make([]models.PartsSelector, 0)

	for i := range stagesInfo.Stages {
		if stagesInfo.Stages[i].Stage == *body.Stage {
			for _, part := range stagesInfo.Stages[i].Parts {
				partsSelector = append(partsSelector, models.PartsSelector{Part: part.Part, PartName: part.PartName})
			}
			break
		}
	}

	utils.BindObjectToRestData(ctx, partsSelector)
}

func (s *ApiService) GetStreetsByStage(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		OfficeID int64 `json:"office_id" binding:"required,OfficeID"`
		WhID     int64 `json:"wh_id" binding:"required,WhID"`
		Stage    int64 `json:"stage" binding:"required,gte=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	streetsInfo, err := s.sprutAccessService.GetStreetsByStage(ctx, employeeID, body.OfficeID, body.WhID, body.Stage)
	if err != nil {
		s.bindServiceError(ctx, err, "can't get streets by stage")
		return
	}

	if len(streetsInfo) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	streetsSelector := make([]models.StreetsSelector, len(streetsInfo))

	for i := range streetsInfo {
		streetsSelector[i] = models.StreetsSelector{
			Street: streetsInfo[i].Street,
		}
	}

	utils.BindObjectToRestData(ctx, streetsSelector)
}

func (s *ApiService) getLocalizedMessage(ctx *gin.Context, dataError rest_data.CustomError) string {
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

func (s *ApiService) bindServiceError(ctx *gin.Context, err error, msg string) {
	if customErr := apierrors.AsCustomError(err); customErr != nil {
		rd := rest_data.RestData{
			Errors:         []rest_data.CustomError{customErr.CustomError},
			HttpResultCode: http.StatusUnprocessableEntity,
		}
		utils.BindRestData(ctx, rd)
		return
	}
	s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("%s: %w", msg, err))
}
