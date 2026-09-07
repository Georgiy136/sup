package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/access"
	custom_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/service/chat_mapper"
	internal_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/utils"
	ticket_validators "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/validators"
	"gitlab.wildberries.ru/wbwh/support/utils.git/access_actions"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

type TicketsService struct {
	validator                    *validator.Validate
	errBuilder                   core_errors.ErrorBuilder
	redisClient                  RedisClientInterface
	resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface
	supportFileManagerApiClient  FileManagerClientInterface
	chatMapper                   *chat_mapper.ChatMapper
	ticketAccess                 *TicketAccessChecker
}

func NewTicketsService(redisClient RedisClientInterface, resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface, supportFileManagerApiClient FileManagerClientInterface, chatMapper *chat_mapper.ChatMapper) *TicketsService {
	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)
	errorsRepo := errors_keys.NewErrorsRepository()
	errorsRepo.SetErrors(support_err_keys.ErrorsRepository)

	return &TicketsService{
		validator:                    valid,
		errBuilder:                   core_errors.NewErrorBuilder(errorsRepo.GetErrors(), nil),
		redisClient:                  redisClient,
		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
		supportFileManagerApiClient:  supportFileManagerApiClient,
		chatMapper:                   chatMapper,
		ticketAccess:                 NewTicketAccessChecker(redisClient, resourceEmployeeAccessClient),
	}
}

func (t TicketsService) CreateTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID      int64           `json:"category_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		TicketName      string          `json:"ticket_name" binding:"required,lte=300"`
		InfoForCreate   json.RawMessage `json:"info_for_create"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.TicketName); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("ticket name must not contain URLs, '%s'", foundURL))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID, employeeID, string(body.InfoForCreate), body.ScenarioOrderID, body.TicketName)
	pg.SetStoredProcedureName("tickets.tickets_create")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) CreateTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID      int64           `json:"category_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		TicketName      string          `json:"ticket_name" binding:"required,lte=300"`
		InfoForCreate   json.RawMessage `json:"info_for_create"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.TicketName); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("ticket name must not contain URLs, '%s'", foundURL))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: "create_ticket_category",
		StatusID:   nil,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for resource: %w", err))
		return
	}

	bodyCreateTicket := bodyCreateTicketWithPublicCategory{
		categoryID:       body.CategoryID,
		employeeID:       employeeID,
		infoForCreate:    body.InfoForCreate,
		scenarioOrderID:  body.ScenarioOrderID,
		ticketName:       body.TicketName,
		isPublicCategory: true,
	}
	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		utils.BindRestData(ctx, createTicketWithPublicCategory(bodyCreateTicket))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for create ticket"))
	}

	bodyCreateTicket.isPublicCategory = false
	utils.BindRestData(ctx, createTicketWithPublicCategory(bodyCreateTicket))
}

func (t TicketsService) CreateTicketV3(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID      int64           `json:"category_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		TicketName      string          `json:"ticket_name" binding:"required,lte=300"`
		Files           json.RawMessage `json:"files"`
		InfoForCreate   json.RawMessage `json:"info_for_create"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.TicketName); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("ticket name must not contain URLs, '%s'", foundURL))
		return
	}

	ticketID, err := reserveTicketID()
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
		return
	}

	var bodyFiles map[string][]models.TicketFiles
	if len(body.Files) > 0 {
		if err := jsoniter.Unmarshal(body.Files, &bodyFiles); err != nil {
			t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't unmarshal files: %w", err))
			return
		}
	}

	bodyCreateTicket := bodyCreateTicketWithReserveTicketID{
		categoryID:      body.CategoryID,
		employeeID:      employeeID,
		infoForCreate:   body.InfoForCreate,
		scenarioOrderID: body.ScenarioOrderID,
		ticketName:      body.TicketName,
		ticketID:        ticketID,
	}
	if len(bodyFiles) == 0 {
		utils.BindRestData(ctx, createTicketWithReserveTicketID(bodyCreateTicket))
		return
	}

	fileIDs := make([]int64, 0)
	for _, fieldFiles := range bodyFiles {
		for _, file := range fieldFiles {
			fileIDs = append(fileIDs, file.FileID)
		}
	}
	if len(fileIDs) > models.MaxFilesPerTicket {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("too many files"))
		return
	}

	fileInfoMap, err := t.getFileInfoMapByID(ctx, employeeID, fileIDs)
	if err != nil {
		if bizErr, ok := errors.AsType[*custom_errors.CustomErrorWrapper](err); ok {
			utils.BindRestData(ctx, *custom_errors.ToRestData(bizErr))
			return
		}
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
		return
	}

	categoryStatusInfo, err := getCategoryStatusInfoForCreate(body.CategoryID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("category statuses info doesn't exist: %w", err))
		return
	}

	allowedFileTypesbyScenarioID, err := getAllowedFileTypesForScenario(categoryStatusInfo, *body.ScenarioOrderID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get allowed file types for category %d: %w", body.CategoryID, err))
		return
	}
	if len(allowedFileTypesbyScenarioID) == 0 {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("no files allowed for ticket category"))
		return
	}

	if err := t.validatedAttachedFiles(body.InfoForCreate, bodyFiles, allowedFileTypesbyScenarioID, fileInfoMap); err != nil {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, err)
		return
	}

	reqAttachFilesToTicket := models.AttachFilesToTicketRequest{
		TicketID: ticketID,
		FileIDs:  fileIDs,
	}
	if err := t.supportFileManagerApiClient.AttachFilesToTicket(ctx, employeeID, reqAttachFilesToTicket); err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't attach files to ticket %d: %w", ticketID, err))
		return
	}

	utils.BindRestData(ctx, createTicketWithReserveTicketID(bodyCreateTicket))
}

func getAllowedFileTypesForScenario(categoryStatusInfo []models.StatusInfo, scenarioOrderID int64) (map[string][]string, error) {
	if len(categoryStatusInfo) == 0 {
		return nil, fmt.Errorf("category ststus info is empty")
	}

	const fileFrontDatatype = "file"
	allowedFileTypesForScenario := make(map[string][]string)

	for _, v := range categoryStatusInfo[0].AddedInformationModel {
		if v.ScenarioOrderID == scenarioOrderID {
			for _, informationModel := range v.ScenarioAddedInformationModel {
				if informationModel.FrontDatatype == fileFrontDatatype {
					allowedFileTypesForScenario[informationModel.DataName] = informationModel.FileAllowedTypes
				}
			}
			break
		}
	}

	return allowedFileTypesForScenario, nil
}

func (t *TicketsService) getFileInfoMapByID(ctx context.Context, employeeID int64, fileIDs []int64) (map[int64]*models.GetFileInfoResponse, error) {
	fileInfoMap := make(map[int64]*models.GetFileInfoResponse, len(fileIDs))
	for _, fileID := range fileIDs {
		req := models.GetFileInfoRequest{
			FileID: fileID,
		}
		fileInfo, err := t.supportFileManagerApiClient.GetFileInfo(ctx, employeeID, req)
		if err != nil {
			return nil, fmt.Errorf("can't get file %d info: %w", fileID, err)

		}

		fileInfoMap[fileID] = fileInfo
	}

	return fileInfoMap, nil
}

func (t TicketsService) validatedAttachedFiles(ext json.RawMessage, bodyFilesMap map[string][]models.TicketFiles, allowedFileTypesForScenario map[string][]string, fileInfoMap map[int64]*models.GetFileInfoResponse) error {
	var extData map[string]json.RawMessage
	if err := jsoniter.Unmarshal(ext, &extData); err != nil {
		return fmt.Errorf("can't parse ext: %w", err)
	}

	for fileFieldName, bodyFiles := range bodyFilesMap {
		extValue, ok := extData[fileFieldName]
		if !ok {
			return fmt.Errorf("file field %s not found in info_for_create", fileFieldName)
		}

		var extFiles []models.TicketFiles
		if err := jsoniter.Unmarshal(extValue, &extFiles); err != nil {
			return fmt.Errorf("can't unmarshal file field %s model from info_for_create", fileFieldName)
		}

		if !slices.Equal(bodyFiles, extFiles) {
			return fmt.Errorf("file field %s model is different in info_for_create and request body", fileFieldName)
		}

		allowedFileTypesForField := internal_utils.SliceToMap(allowedFileTypesForScenario[fileFieldName])
		for _, file := range bodyFiles {
			if _, ok := allowedFileTypesForField[file.FileType]; !ok {
				return fmt.Errorf("type %s is not allowed for field %s, allowed: %v", file.FileType, fileFieldName, allowedFileTypesForField)
			}

			fileInfo, ok := fileInfoMap[file.FileID]
			if !ok {
				return fmt.Errorf("file %d doesn't exist in db", file.FileID)
			}
			if err := compareFileFields(file, fileInfo); err != nil {
				return fmt.Errorf("file %d compare field error: %w", file.FileID, err)
			}
		}
	}

	return nil
}

func compareFileFields(file models.TicketFiles, fileInfo *models.GetFileInfoResponse) error {
	switch {
	case file.FileSize != fileInfo.SizeBytes:
		return fmt.Errorf("size mismatch in file %d; expected: %d, actual: %d", file.FileID, fileInfo.SizeBytes, file.FileSize)
	case file.FileName != fileInfo.OriginalName:
		return fmt.Errorf("name mismatch in file %d; expected: %s, actual: %s", file.FileID, fileInfo.OriginalName, file.FileName)
	case file.FileType != fileInfo.MimeType:
		return fmt.Errorf("type mismatch in file %d; expected: %s, actual: %s", file.FileID, fileInfo.MimeType, file.FileType)
	}

	return nil
}

func (t TicketsService) PerformTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		Ext             json.RawMessage `json:"ext"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	switch body.Ext {
	case nil:
		pg.SetParams(body.TicketID, nil, employeeID, body.ScenarioOrderID)
	default:
		pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID)
	}
	pg.SetStoredProcedureName("tickets.tickets_perform")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) PerformTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		Ext             json.RawMessage `json:"ext"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID)
	pg.SetStoredProcedureName("tickets.tickets_perform_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) PerformTicketV3(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		Ext             json.RawMessage `json:"ext"`
		ExtVersion      *string         `json:"ext_version" binding:"omitempty,uuid4"`
		Files           json.RawMessage `json:"files"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if len(body.Files) > 0 {
		var parsedFiles map[string][]models.TicketFiles
		if err := jsoniter.Unmarshal(body.Files, &parsedFiles); err != nil {
			t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
			return
		}
		if len(parsedFiles) > 0 {
			ticketCommonInfo, err := getTicketCommonInfo(body.TicketID)
			if err != nil {
				t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("get ticket common info error: %w", err))
				return
			}
			if ticketCommonInfo == nil {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("ticket not found"))
				return
			}

			fileIDs := make([]int64, 0, len(parsedFiles))
			for _, fieldFiles := range parsedFiles {
				for _, file := range fieldFiles {
					fileIDs = append(fileIDs, file.FileID)
				}
			}
			if len(fileIDs) > models.MaxFilesPerTicket {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("too many files, got %d, max %d", len(fileIDs), models.MaxFilesPerTicket))
				return
			}

			fileInfoMap, err := t.getFileInfoMapByID(ctx, employeeID, fileIDs)
			if err != nil {
				if bizErr, ok := errors.AsType[*custom_errors.CustomErrorWrapper](err); ok {
					utils.BindRestData(ctx, *custom_errors.ToRestData(bizErr))
					return
				}
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
				return
			}

			categoryStatusInfo, err := getCategoryStatusInfo(ticketCommonInfo.CategoryID, ticketCommonInfo.StatusID)
			if err != nil {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("category statuses info doesn't exist: %w", err))
				return
			}

			allowedFileTypesbyScenarioID, err := getAllowedFileTypesForScenario(categoryStatusInfo, *body.ScenarioOrderID)
			if err != nil {
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get allowed file types for category %d: %w", ticketCommonInfo.CategoryID, err))
				return
			}
			if len(allowedFileTypesbyScenarioID) == 0 {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("no files allowed for ticket category"))
				return
			}

			if err := t.validatedAttachedFiles(body.Ext, parsedFiles, allowedFileTypesbyScenarioID, fileInfoMap); err != nil {
				t.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, err)
				return
			}

			reqAttachFilesToTicket := models.AttachFilesToTicketRequest{
				TicketID: body.TicketID,
				FileIDs:  fileIDs,
			}
			if err := t.supportFileManagerApiClient.AttachFilesToTicket(ctx, employeeID, reqAttachFilesToTicket); err != nil {
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't attach files to ticket %d: %w", body.TicketID, err))
				return
			}
		}
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	switch body.Ext {
	case nil:
		pg.SetParams(body.TicketID, nil, employeeID, body.ScenarioOrderID, nil, nil, body.ExtVersion)
	default:
		pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID, nil, nil, body.ExtVersion)
	}
	pg.SetStoredProcedureName("tickets.tickets_perform")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) RejectTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID         int64  `json:"ticket_id" binding:"required,gt=0"`
		RejectedComments string `json:"rejected_comments" binding:"required,max=300"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.RejectedComments); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("rejected comments must not contain URLs, '%s'", foundURL))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.RejectedComments)
	pg.SetStoredProcedureName("tickets.tickets_reject")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) RejectTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID         int64  `json:"ticket_id" binding:"required,gt=0"`
		RejectedComments string `json:"rejected_comments" binding:"required,max=300"`
		RejectedFile     []struct {
			FileID   int64  `json:"file_id" binding:"required,gt=0"`
			FileType string `json:"file_type" binding:"required,min=1,max=100"`
			FileSize int64  `json:"file_size" binding:"required,gt=0"`
			FileName string `json:"file_name" binding:"required,min=1,max=100"`
		} `json:"rejected_file" binding:"omitempty,gt=0"`
		ExtVersion *string `json:"ext_version" binding:"omitempty,uuid4"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.RejectedComments); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("rejected comments must not contain URLs, '%s'", foundURL))
		return
	}

	if len(body.RejectedFile) > 0 {
		fileIDs := make([]int64, 0, len(body.RejectedFile))

		for _, file := range body.RejectedFile {
			req := models.GetFileInfoRequest{
				FileID: file.FileID,
			}

			fileInfo, err := t.supportFileManagerApiClient.GetFileInfo(ctx, employeeID, req)
			if err != nil {
				if bizErr, ok := errors.AsType[*custom_errors.CustomErrorWrapper](err); ok {
					utils.BindRestData(ctx, *custom_errors.ToRestData(bizErr))
					return
				}
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get file %d info: %w", file.FileID, err))
				return
			}

			f := models.TicketFiles{
				FileID:   file.FileID,
				FileType: file.FileType,
				FileSize: file.FileSize,
				FileName: file.FileName,
			}
			if err := compareFileFields(f, fileInfo); err != nil {
				t.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, fmt.Errorf("file %d compare field error: %w", file.FileID, err))
				return
			}

			fileIDs = append(fileIDs, file.FileID)
		}

		req := models.AttachFilesToTicketRequest{
			TicketID: body.TicketID,
			FileIDs:  fileIDs,
		}
		if err := t.supportFileManagerApiClient.AttachFilesToTicket(ctx, employeeID, req); err != nil {
			t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't attach files to ticket %d: %w", body.TicketID, err))
			return
		}
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.RejectedComments, nil, nil, nil, body.RejectedFile, body.ExtVersion)
	pg.SetStoredProcedureName("tickets.tickets_reject")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketsCreatedByEmployee(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.tickets_getallbyemployee")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapCreatedByEmployeeTicketsWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetTicketsCreatedByEmployee] MapCreatedByEmployeeTicketsWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64  `json:"ticket_id" binding:"required,gt=0"`
		Tab      string `json:"tab" binding:"required,min=1,max=30"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.Tab)
	pg.SetStoredProcedureName("tickets.tickets_getbyticket")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketsForWorkV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.ticketsemployee_getallforwork")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapEmployeeWorkTicketsWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetTicketsForWorkV2] MapEmployeeWorkTicketsWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) ApproveTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		Ext             json.RawMessage `json:"ext"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	switch body.Ext {
	case nil:
		pg.SetParams(body.TicketID, nil, employeeID, body.ScenarioOrderID)
	default:
		pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID)
	}
	pg.SetStoredProcedureName("tickets.tickets_approve")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) ApproveTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction      string          `json:"type_action" binding:"required,min=3,max=50"`
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		CategoryID      int64           `json:"category_id" binding:"required,gt=0"`
		StatusID        *string         `json:"status_id" binding:"omitempty,min=1,max=3"`
		Ext             json.RawMessage `json:"ext"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access policy for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no access policy found for ticket: %d, category: %d, type_action: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no employee access for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID)
	pg.SetStoredProcedureName("tickets.tickets_approve_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) ApproveTicketV3(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID        int64           `json:"ticket_id" binding:"required,gt=0"`
		ScenarioOrderID *int64          `json:"scenario_order_id" binding:"required,gte=0"`
		Ext             json.RawMessage `json:"ext"`
		ExtVersion      *string         `json:"ext_version" binding:"omitempty,uuid4"`
		Files           json.RawMessage `json:"files"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if len(body.Files) > 0 {
		var parsedFiles map[string][]models.TicketFiles
		if err := jsoniter.Unmarshal(body.Files, &parsedFiles); err != nil {
			t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
			return
		}
		if len(parsedFiles) > 0 {
			ticketCommonInfo, err := getTicketCommonInfo(body.TicketID)
			if err != nil {
				t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("get ticket common info error: %w", err))
				return
			}
			if ticketCommonInfo == nil {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("ticket not found"))
				return
			}

			fileIDs := make([]int64, 0, len(parsedFiles))
			for _, fieldFiles := range parsedFiles {
				for _, file := range fieldFiles {
					fileIDs = append(fileIDs, file.FileID)
				}
			}
			if len(fileIDs) > models.MaxFilesPerTicket {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("too many files, got %d, max %d", len(fileIDs), models.MaxFilesPerTicket))
				return
			}

			fileInfoMap, err := t.getFileInfoMapByID(ctx, employeeID, fileIDs)
			if err != nil {
				if bizErr, ok := errors.AsType[*custom_errors.CustomErrorWrapper](err); ok {
					utils.BindRestData(ctx, *custom_errors.ToRestData(bizErr))
					return
				}
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
				return
			}

			categoryStatusInfo, err := getCategoryStatusInfo(ticketCommonInfo.CategoryID, ticketCommonInfo.StatusID)
			if err != nil {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("category statuses info doesn't exist: %w", err))
				return
			}

			allowedFileTypesbyScenarioID, err := getAllowedFileTypesForScenario(categoryStatusInfo, *body.ScenarioOrderID)
			if err != nil {
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get allowed file types for category %d: %w", ticketCommonInfo.CategoryID, err))
				return
			}
			if len(allowedFileTypesbyScenarioID) == 0 {
				t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("no files allowed for ticket category"))
				return
			}

			if err := t.validatedAttachedFiles(body.Ext, parsedFiles, allowedFileTypesbyScenarioID, fileInfoMap); err != nil {
				t.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, err)
				return
			}

			reqAttachFilesToTicket := models.AttachFilesToTicketRequest{
				TicketID: body.TicketID,
				FileIDs:  fileIDs,
			}
			if err := t.supportFileManagerApiClient.AttachFilesToTicket(ctx, employeeID, reqAttachFilesToTicket); err != nil {
				t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't attach files to ticket %d: %w", body.TicketID, err))
				return
			}
		}
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	switch body.Ext {
	case nil:
		pg.SetParams(body.TicketID, nil, employeeID, body.ScenarioOrderID, body.ExtVersion)
	default:
		pg.SetParams(body.TicketID, string(body.Ext), employeeID, body.ScenarioOrderID, body.ExtVersion)
	}
	pg.SetStoredProcedureName("tickets.tickets_approve")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) BookTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID   int64   `json:"ticket_id" binding:"required,gt=0"`
		ExtVersion *string `json:"ext_version" binding:"omitempty,uuid4"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.ExtVersion)
	pg.SetStoredProcedureName("tickets.tickets_booking")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) BookTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction string  `json:"type_action" binding:"required,min=3,max=50"`
		TicketID   int64   `json:"ticket_id" binding:"required,gt=0"`
		CategoryID int64   `json:"category_id" binding:"required,gt=0"`
		StatusID   *string `json:"status_id" binding:"omitempty,min=1,max=3"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access policy for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no access policy found for ticket: %d, category: %d, type_action: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no employee access for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID)
	pg.SetStoredProcedureName("tickets.tickets_booking_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) UnbookTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID   int64   `json:"ticket_id" binding:"required,gt=0"`
		ExtVersion *string `json:"ext_version" binding:"omitempty,uuid4"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.ExtVersion)
	pg.SetStoredProcedureName("tickets.tickets_unbooking")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) UnbookTicketV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID)
	pg.SetStoredProcedureName("tickets.tickets_unbooking_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) UpdateFavouriteTickets(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		utils.BindValidationErrorWithAbort(ctx, "can't parse ch employee id parameter")
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID         int64 `json:"ticket_id" binding:"required"`
		EmployeeID       int64 `json:"employee_id" binding:"required,EmployeeID"`
		NeedNotification *bool `json:"need_notification" binding:"required"`
		IsDel            *bool `json:"is_del" binding:"required"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, body.EmployeeID, body.NeedNotification, chEmployeeID, body.IsDel)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_updfavourite")

	rd := repository.GetRestDataFromDb(pg)

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) TicketsFavouriteGetByEmployeeV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_getallbyemployee")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapTicketListWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[TicketsFavouriteGetByEmployeeV2] MapTicketListWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) TicketsFavouriteGetByEmployeeForWorkV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_getallbyemployeeforwork")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapTicketListWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[TicketsFavouriteGetByEmployeeForWorkV2] MapTicketListWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketInfoByTypeAction(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID   int64   `json:"ticket_id" binding:"required,gt=0"`
		CategoryID int64   `json:"category_id" binding:"required,gt=0"`
		TypeAction string  `json:"type_action" binding:"required,min=3,max=50"`
		StatusID   *string `json:"status_id" binding:"omitempty,min=1,max=3"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no actions found for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no employee access for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, []int64{body.CategoryID})
	pg.SetStoredProcedureName("tickets.tickets_getforcenter")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketInfoForCenter(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID   int64   `json:"ticket_id" binding:"required,gt=0"`
		CategoryID int64   `json:"category_id" binding:"required,gt=0"`
		TypeAction string  `json:"type_action" binding:"required,min=3,max=50"`
		StatusID   *string `json:"status_id" binding:"omitempty,min=1,max=3"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no actions found for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no employee access for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, []int64{body.CategoryID})
	pg.SetStoredProcedureName("tickets.tickets_getforcenter")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketInfoForMyTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID)
	pg.SetStoredProcedureName("tickets.tickets_getforown")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) ShareMyTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		RecipientEmployeeID int64 `json:"recipient_employee_id" binding:"required,EmployeeID"`
		TicketID            int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, body.RecipientEmployeeID, employeeID)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_addemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) ShareWorkTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		RecipientEmployeeID int64   `json:"recipient_employee_id" binding:"required,EmployeeID"`
		TypeAction          string  `json:"type_action" binding:"required,oneof=status_approve_category status_perform_category"`
		TicketID            int64   `json:"ticket_id" binding:"required,gt=0"`
		CategoryID          int64   `json:"category_id" binding:"required,gt=0"`
		StatusID            *string `json:"status_id" binding:"omitempty,min=1,max=3"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access policy for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no access policy found for ticket: %d, category: %d, type_action: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no employee access for ticket: %d, category: %d, action type: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, body.RecipientEmployeeID, employeeID, true)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_addemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) DeleteTicketFromObservables(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID)
	pg.SetStoredProcedureName("tickets.ticketsfavourite_remotemployee")

	rd := repository.GetRestDataFromDb(pg)

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetWorkTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	categoryStatusAccess, err := t.ticketAccess.GetEmployeeCategoryStatusAccess(ctx.Request.Context(), employeeID, access_actions.TypeActionStatusApproveCategory, access_actions.TypeActionStatusPerformCategory)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, err)
		return
	}
	if len(categoryStatusAccess) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	categoryStatusIDsJSON, err := jsoniter.Marshal(categoryStatusAccess)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal category status access: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(string(categoryStatusIDsJSON), employeeID)
	pg.SetStoredProcedureName("tickets.tickets_getallforwork")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapTicketListWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetWorkTickets] MapTicketListWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) GetTicketForLink(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, err := t.ticketAccess.HasEmployeeTicketAccess(ctx, employeeID, body.TicketID,
		access_actions.TypeActionCreateTicketCategory,
		access_actions.TypeActionViewTicketCategory,
		access_actions.TypeActionStatusApproveCategory,
		access_actions.TypeActionStatusPerformCategory,
	)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTicketNotFound) {
			t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketNotFound, err)
			return
		}
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("check ticket access error: %w", err))
		return
	}
	if !hasAccess {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, errors.New("no access to get info ticket"))
		return
	}

	employeeGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, employeeGroups)
	pg.SetStoredProcedureName("tickets.tickets_getbyticketforlink")
	utils.BindRestData(ctx, repository.GetRestDataFromDb(pg))
}

func (t TicketsService) GetMyTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.tickets_getallforown")
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := t.chatMapper.MapOwnTicketsWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetMyTickets] MapOwnTicketsWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (t TicketsService) RejectMyTicketOrTicketAtPerform(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID         int64  `json:"ticket_id" binding:"required,gt=0"`
		RejectedComments string `json:"rejected_comments" binding:"required,max=400"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.RejectedComments); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("rejected comments must not contain URLs, '%s'", foundURL))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.RejectedComments)
	pg.SetStoredProcedureName("tickets.tickets_reject_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) RejectTicketAtApproveOrBook(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction       string  `json:"type_action" binding:"required,oneof=status_approve_category status_perform_category"`
		TicketID         int64   `json:"ticket_id" binding:"required,gt=0"`
		CategoryID       int64   `json:"category_id" binding:"required,gt=0"`
		StatusID         *string `json:"status_id" binding:"omitempty,min=1,max=3"`
		RejectedComments string  `json:"rejected_comments" binding:"required,max=400"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.RejectedComments); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("rejected comments must not contain URLs, '%s'", foundURL))
		return
	}

	resourceAccessPolicy, err := t.redisClient.GetResourceAccessPolicy(ctx, models.ResourceCompositeKey{
		CategoryID: body.CategoryID,
		TypeAction: body.TypeAction,
		StatusID:   body.StatusID,
	})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access policy for resource: %w", err))
		return
	}

	if access.IsAccessPolicyEmpty(resourceAccessPolicy) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"no access policy found for ticket: %d, category: %d, type_action: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	employeeAccessGroups, err := t.redisClient.GetAccessGroupsByEmployee(ctx, employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access groups for employee: %w", err))
		return
	}

	employeeExternalActions, err := t.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee: %w", err))
		return
	}

	if !access.HasAccessByPolicy(*resourceAccessPolicy, employeeAccessGroups, employeeExternalActions) {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf(
			"employee has no access for ticket: %d, category: %d, type_action: %s, status: %v",
			body.TicketID,
			body.CategoryID,
			body.TypeAction,
			body.StatusID,
		))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.RejectedComments, nil, true)
	pg.SetStoredProcedureName("tickets.tickets_reject_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t TicketsService) RequestToRejectTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID         int64  `json:"ticket_id" binding:"required,gt=0"`
		RejectedComments string `json:"rejected_comments" binding:"required,max=400"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	if foundURL, hasURL := ticket_validators.FindURL(body.RejectedComments); hasURL {
		t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketContainUrl, fmt.Errorf("rejected comments must not contain URLs, '%s'", foundURL))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, body.RejectedComments)
	pg.SetStoredProcedureName("tickets.tickets_prereject")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t *TicketsService) UpdateTicketExt(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64           `json:"ticket_id" binding:"required,gt=0"`
		Ext      json.RawMessage `json:"ext"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, string(body.Ext), employeeID)
	pg.SetStoredProcedureName("tickets.tickets_ticketsextupd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t *TicketsService) GetTicketFieldInfoByCategoryStatus(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID int64   `json:"category_id" binding:"required,gt=0"`
		StatusID   *string `json:"status_id" binding:"omitempty,min=1,max=3"`
		DataName   string  `json:"data_name" binding:"required,lte=30"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID, body.StatusID, body.DataName)
	pg.SetStoredProcedureName("tickets.categorystatus_getbyext")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (t *TicketsService) ReturnTicketStatus(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID       int64  `json:"ticket_id" binding:"required,gt=0"`
		ReturnStatusID string `json:"return_status_id" binding:"required,min=1,max=3"`
		ReturnComment  string `json:"return_comment" binding:"required,max=500"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, body.ReturnStatusID, body.ReturnComment, employeeID)
	pg.SetStoredProcedureName("tickets.tickets_ticketsreturnstatus")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
