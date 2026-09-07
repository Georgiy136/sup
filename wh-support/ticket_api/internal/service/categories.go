package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

type CategoryInfo struct {
	validator  *validator.Validate
	errBuilder core_errors.ErrorBuilder
}

func NewCategoryInfo() *CategoryInfo {
	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)

	return &CategoryInfo{
		validator:  valid,
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (c CategoryInfo) GetCategoryInfoByIDV2(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID int64   `json:"category_id" binding:"required,gt=0"`
		StatusID   *string `json:"status_id" binding:"omitempty,gt=0,max=3"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	switch body.StatusID {
	case nil:
		pg.SetParams(body.CategoryID)
	default:
		pg.SetParams(body.CategoryID, body.StatusID)
	}
	pg.SetStoredProcedureName("tickets.categorystatus_getbystatus")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) AddCategory(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.RawBODY].([]byte)
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if len(body) == 0 {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	var bodyStruct models.AddCategoryRequest
	if err := jsoniter.Unmarshal(body, &bodyStruct); err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500UnmarshalWrong, fmt.Errorf("can't unmarshal body: %w", err))
		return
	}
	if err := c.validator.Struct(bodyStruct); err != nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("validate body error: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(bodyStruct.CategoryName, bodyStruct.ParentCategoryID, bodyStruct.ActionID, chEmployeeID, *bodyStruct.IsActive)

	pg.SetStoredProcedureName("tickets.category_add")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) UpdateCategoryStatus(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID int64 `json:"category_id" binding:"required"`
		StatusInfo []struct {
			IsMain                *bool           `json:"is_main" binding:"required"`
			StatusID              string          `json:"status_id" binding:"required"`
			ConstructorOrderID    int64           `json:"constructor_order_id" binding:"required"`
			ApproveGroupID        *int64          `json:"approve_group_id" binding:"omitempty"`
			PerformGroupID        *int64          `json:"perform_group_id" binding:"omitempty"`
			HasPrereject          *bool           `json:"has_prereject" binding:"required"`
			InformationModel      json.RawMessage `json:"information_model" binding:"omitempty"`
			ApproveGroupName      *string         `json:"approve_group_name" binding:"omitempty,lte=100"`
			PerformGroupName      *string         `json:"perform_group_name" binding:"omitempty,lte=100"`
			StatusDescription     string          `json:"status_description" binding:"required,lte=100"`
			AutoAPI               *string         `json:"autoapi"`
			ScenarioLabel         *string         `json:"scenario_label" binding:"omitempty,lte=100"`
			WarningDescription    *string         `json:"warning_description" binding:"omitempty,lte=100"`
			InfoDescription       *string         `json:"info_description" binding:"omitempty,lte=100"`
			IsBlockedReject       *bool           `json:"is_blocked_reject" binding:"omitempty"`
			ReturnStatusIDs       []string        `json:"return_status_ids" binding:"omitempty,dive,min=1,max=3"`
			IsBlockedStatus       *bool           `json:"is_blocked_status" binding:"omitempty"`
			AddedInformationModel []struct {
				ScenarioOrderID               *int64 `json:"scenario_order_id" binding:"required,gte=0"`
				ScenarioName                  string `json:"scenario_name" binding:"required,lte=100"`
				NextStatusID                  string `json:"next_status_id" binding:"required,lte=3"`
				ScenarioAddedInformationModel []struct {
					DataName           string          `json:"data_name" binding:"required,lte=30"`
					OrderID            int64           `json:"order_id" binding:"required"`
					DataType           string          `json:"data_type" binding:"required,lte=15"`
					FrontDataName      string          `json:"front_data_name" binding:"required,lte=100"`
					IsRequired         *bool           `json:"is_required" binding:"required"`
					FrontModel         json.RawMessage `json:"front_model" binding:"required"`
					IsHiddenForCreator *bool           `json:"is_hidden_for_creator" binding:"required"`
				} `json:"scenario_added_information_model" binding:"required"`
			} `json:"added_information_model" binding:"omitempty"`
		} `json:"status_info" binding:"required"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	statusInfo, err := jsoniter.Marshal(body.StatusInfo)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal body: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID, string(statusInfo), chEmployeeID)
	pg.SetStoredProcedureName("tickets.categorystatus_upd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) GetCategoriesTree(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.category_getalltree")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) GetCategoryStatusesInfo(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID int64 `json:"category_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID)
	pg.SetStoredProcedureName("tickets.categorystatus_getbycategory")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) UpdateCategory(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.RawBODY].([]byte)
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if len(body) == 0 {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	var bodyStruct models.UpdateCategoryRequest
	if err := jsoniter.Unmarshal(body, &bodyStruct); err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500UnmarshalWrong, fmt.Errorf("can't unmarshal body: %w", err))
		return
	}
	if err := c.validator.Struct(bodyStruct); err != nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("validate body error: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(bodyStruct.CategoryID, bodyStruct.CategoryName, bodyStruct.ParentCategoryID, bodyStruct.ActionID, chEmployeeID, *bodyStruct.IsDel, bodyStruct.Icon, bodyStruct.Color)
	pg.SetStoredProcedureName("tickets.category_upd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) CopyCategoryTicket(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID       int64 `json:"category_id" binding:"required"`
		ParentCategoryID int64 `json:"parent_category_id" binding:"required"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID, body.ParentCategoryID, chEmployeeID)
	pg.SetStoredProcedureName("tickets.category_copy")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) CopyCategorySettings(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		SourceCategoryID int64 `json:"source_category_id" binding:"required,gt=0"`
		TargetCategoryID int64 `json:"target_category_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.SourceCategoryID, body.TargetCategoryID, chEmployeeID)
	pg.SetStoredProcedureName("tickets.category_settingscopy")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c CategoryInfo) MoveCategory(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID       int64 `json:"category_id" binding:"required,gt=0"`
		ParentCategoryID int64 `json:"parent_category_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.CategoryID, body.ParentCategoryID, chEmployeeID)
	pg.SetStoredProcedureName("tickets.category_move")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
