package service

import (
	"context"
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/models"
	localutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/utils"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type resourcesService struct {
	redisClient            *clients.RedisClient
	employeeInfoApiClient  *clients.EmployeeInfoApiClient
	accessManagerApiClient *clients.AccessManagerApiClient
	errBuilder             core_errors.ErrorBuilder
}

func NewResourcesService() *resourcesService {
	return &resourcesService{

		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (r *resourcesService) Init(ctx context.Context, cfg configs.Config) {
	r.redisClient = clients.NewRedisClient(ctx, cfg)
	r.employeeInfoApiClient = clients.NewEmployeeInfoApiClient(ctx, cfg)
	r.accessManagerApiClient = clients.NewAccessManagerApiClient(ctx, cfg)
}

func (r *resourcesService) GetResourcesHandler(ctx *gin.Context, _ map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)
	result, err := r.redisClient.GetFromCache(ctx, employeeID)
	if err != nil {
		utils.BindServiceErrorWithAbort(ctx, "can't get resources from cache", err)
		return
	}
	if result != nil {
		utils.BindObjectToRestData(ctx, models.EmployeeResourcesResp{ActionIds: result.ActionIds})
		return
	}
	utils.BindNoContent(ctx)
}

func (r *resourcesService) GetAllActionsHandler(ctx *gin.Context, _ map[string]interface{}) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetStoredProcedureName("access.actions_getall")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) CreateActionHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		ActionID   string `json:"action_id" binding:"required,max=50"`
		ActionDesc string `json:"action_desc" binding:"required,max=200"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.ActionID, body.ActionDesc, employeeID)
	pg.SetStoredProcedureName("access.actions_addaction")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) UpdateActionHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		EmployeeID int64  `json:"employee_id" binding:"required,EmployeeID"`
		ActionID   string `json:"action_id" binding:"required,max=50"`
		IsDel      *bool  `json:"is_del" binding:"omitempty"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.EmployeeID, body.ActionID, employeeID, body.IsDel)
	pg.SetStoredProcedureName("access.employeeactions_upd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) DeleteActionHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	actionID, ok := params["action_id"].(string)
	if !ok || actionID == "" {
		utils.BindValidationErrorWithAbort(ctx, "can't parse actionID parameter")
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(actionID, employeeID)
	pg.SetStoredProcedureName("access.actions_deleteaction")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) GetEmployeesByActionHandler(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		ActionID string `json:"action_id" binding:"required,max=50"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.ActionID)
	pg.SetStoredProcedureName("access.employeeactions_getbyaction")

	data := repository.GetRestDataFromDb(pg)
	switch {
	case data.HasError():
		logrus.Warnf("can't get data from db by group %s: %v", body.ActionID, data.Error())
		utils.BindRestData(ctx, data)
		return
	case data.DataEmpty():
		logrus.Debugf("empty data from db by group %s", body.ActionID)
		utils.BindNoContent(ctx)
		return
	}

	var res models.DBEmployeesByAction
	if err := jsoniter.Unmarshal(data.GetData(), &res); err != nil {
		logrus.Warnf("can't unmarshal employee info from db: %v", err)
		r.errBuilder.BindError(ctx, errors_keys.Err500UnmarshalWrong, fmt.Errorf("can't unmarshal employee info: %w", err))
		return
	}

	if res.EmployeeInfo == nil || len(res.EmployeeInfo) == 0 {
		logrus.Debugf("list employees id on action %s is empty", body.ActionID)
		utils.BindNoContent(ctx)
		return
	}

	employeeIDs := localutils.DedupEmployeeIDsFromEmployeeInfoFromDB(res.EmployeeInfo)

	employeeInfo, err := r.employeeInfoApiClient.GetEmployeeInfo(chEmployeeID, employeeIDs)
	if err != nil {
		logrus.Errorf("can't get employee info: %v", err)
	}

	if employeeInfo == nil {
		logrus.Warnf("nil from employee info: %v", common.ErrEmployeeUnknown)
	}

	employeeNameMap := r.getEmployeeNameMap(employeeInfo)

	resp := make([]models.EmployeeInfoResp, 0, len(res.EmployeeInfo))
	for i := range res.EmployeeInfo {
		resp = append(resp, models.EmployeeInfoResp{
			ID:             res.EmployeeInfo[i].EmployeeId,
			ChEmployeeID:   res.EmployeeInfo[i].ChEmployeeID,
			ChDt:           res.EmployeeInfo[i].ChDt,
			Name:           employeeNameMap[res.EmployeeInfo[i].EmployeeId],
			ChEmployeeName: employeeNameMap[res.EmployeeInfo[i].ChEmployeeID],
		})
	}

	utils.BindObjectToRestData(ctx, &resp)
}

func (r *resourcesService) GetAccessActionsByEmployeeIDHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	result, err := r.redisClient.GetFromCacheByExternalID(ctx, employeeID)
	if err != nil {
		r.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get resources from cache: %w", err))
		return
	}
	if result != nil {
		utils.BindObjectToRestData(ctx, models.EmployeeResourcesResp{ActionIds: result})
		return
	}

	employeeAppActions, err := r.accessManagerApiClient.GetPermittedAppActions(employeeID)
	if err != nil {
		r.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get employee app actions info: %w", err))
		return
	}

	actions := make([]string, len(employeeAppActions))
	for i := range employeeAppActions {
		actions[i] = employeeAppActions[i].AppActionName
	}

	go func(employeeID int64, actions []string) {
		if err := r.redisClient.SaveToCacheByExternalID(context.Background(), employeeID, actions); err != nil {
			logrus.Errorf("can't save resource to cache: %v", err)
		}
	}(employeeID, actions)

	utils.BindObjectToRestData(ctx, models.EmployeeResourcesResp{ActionIds: actions})
}

func (r *resourcesService) getEmployeeNameMap(info []models.EmployeeInfo) map[int64]*string {
	employeeNameMap := make(map[int64]*string, len(info))

	for i := range info {
		employeeNameMap[info[i].EmployeeId] = &info[i].Name
	}

	return employeeNameMap
}

func (r *resourcesService) GetTypeActionsHandler(ctx *gin.Context, _ map[string]interface{}) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetStoredProcedureName("access.typeactions_getall")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) AddTypeActionHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeActionID   string `json:"typeaction_id" binding:"required,min=5,max=50"`
		TypeActionDesc string `json:"typeaction_desc" binding:"required,min=5,max=200"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.TypeActionID, body.TypeActionDesc, employeeID)
	pg.SetStoredProcedureName("access.typeactions_addtypeaction")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (r *resourcesService) DeleteTypeActionHandler(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		r.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeActionID string `json:"typeaction_id" binding:"required,min=5,max=50"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.TypeActionID, employeeID)
	pg.SetStoredProcedureName("access.typeactions_deletetypeaction")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
