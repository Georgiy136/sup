package service

import (
	"context"
	"errors"
	"fmt"

	client "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/converters"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/models"
	customutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/utils"
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

type EmployeeGroupService struct {
	employeeInfoClient *client.EmployeeInfoApiClient
	errBuilder         core_errors.ErrorBuilder
}

func NewPersonalAccount(c *client.EmployeeInfoApiClient) *EmployeeGroupService {
	return &EmployeeGroupService{
		employeeInfoClient: c,
		errBuilder:         core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (p *EmployeeGroupService) Configure(ctx context.Context, configs configs.Config) {
	p.employeeInfoClient.Configure(ctx, configs)
}

func (p *EmployeeGroupService) CreateGroup(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupName      string  `json:"group_name" binding:"required,max=100"`
		ExternalAction *string `json:"external_action" binding:"omitempty,max=100"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupName, employeeID, body.ExternalAction)
	pg.SetStoredProcedureName("hr.group_create")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) UpdateExternalActionGroup(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID        int64   `json:"group_id" binding:"required,LteInt32"`
		ExternalAction *string `json:"external_action" binding:"omitempty,max=100"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID, body.ExternalAction, chEmployeeID)
	pg.SetStoredProcedureName("hr.group_externalactionupd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) DeleteGroup(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID int64 `json:"group_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID, employeeID)
	pg.SetStoredProcedureName("hr.group_delete")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) GetEmployeesByGroup(ctx *gin.Context, params map[string]interface{}) {
	chEmployeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID int64 `json:"group_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID)
	pg.SetStoredProcedureName("hr.employeegroups_getbygroup")

	data := repository.GetRestDataFromDb(pg)
	switch {
	case data.HasError():
		logrus.Warnf("can't get data from db by group %d: %v", body.GroupID, data.Error())
		utils.BindRestData(ctx, data)
		return
	case data.DataEmpty():
		logrus.Debugf("empty data from db by group %d", body.GroupID)
		utils.BindNoContent(ctx)
		return
	}

	var dbResp []models.DBEmployeesByGroup

	err := jsoniter.Unmarshal(data.GetData(), &dbResp)
	if err != nil {
		logrus.Warnf("can't unmarshal employee info from db: %v", err)
		p.errBuilder.BindError(ctx, errors_keys.Err500UnmarshalWrong, fmt.Errorf("can't unmarshal employee info: %w", err))
		return
	}

	if len(dbResp[0].EmployeeInfo) == 0 {
		logrus.Debugf("group without employes %d", body.GroupID)
		utils.BindNoContent(ctx)
		return
	}

	employeesID := make([]int64, 0, 2*len(dbResp[0].EmployeeInfo))

	for _, employee := range dbResp[0].EmployeeInfo {
		employeesID = append(employeesID, employee.EmployeeID, employee.ChEmployeeID)
	}

	reqForGetEmployeesFullName := customutils.DeduplicateNumsSlice(employeesID)

	employeesNameInfo, err := p.employeeInfoClient.GetEmployeesFullName(chEmployeeID, reqForGetEmployeesFullName)
	if err != nil {
		logrus.Errorf("can't get employee info: %v", err)
	}

	employeesMap := converters.ConvertEmployeeInfoInMap(employeesNameInfo)

	resp := make([]models.EmployeeInfoResp, len(dbResp[0].EmployeeInfo))

	for i := range dbResp[0].EmployeeInfo {
		resp[i] = models.EmployeeInfoResp{
			ChDt:         dbResp[0].EmployeeInfo[i].ChDt,
			EmployeeID:   dbResp[0].EmployeeInfo[i].EmployeeID,
			ChEmployeeID: dbResp[0].EmployeeInfo[i].ChEmployeeID,
		}

		if employeeName, ok := employeesMap[dbResp[0].EmployeeInfo[i].EmployeeID]; ok {
			resp[i].EmployeeName = &employeeName
		}

		if employeeName, ok := employeesMap[dbResp[0].EmployeeInfo[i].ChEmployeeID]; ok {
			resp[i].ChEmployeeName = &employeeName
		}

	}

	utils.BindObjectToRestData(ctx, &resp)
}

func (p *EmployeeGroupService) GetGroupsByEmployee(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		EmployeeID int64 `json:"employee_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.EmployeeID)
	pg.SetStoredProcedureName("hr.employeegroups_getbyemployee")

	data := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, data)
}

func (p *EmployeeGroupService) GetAllGroups(ctx *gin.Context, _ map[string]interface{}) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetStoredProcedureName("hr.group_getall")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) AddEmployeeToGroup(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID    int64 `json:"group_id" binding:"required,LteInt32"`
		EmployeeID int64 `json:"employee_id" binding:"required,EmployeeID"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID, body.EmployeeID, employeeID)
	pg.SetStoredProcedureName("hr.group_addemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) DeleteEmployeeFromGroup(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID    int64 `json:"group_id" binding:"required,LteInt32"`
		EmployeeID int64 `json:"employee_id" binding:"required,EmployeeID"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID, body.EmployeeID, employeeID)
	pg.SetStoredProcedureName("hr.group_removeemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) GetCategoriesWithGroups(ctx *gin.Context, params map[string]interface{}) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetStoredProcedureName("tickets.categorygroup_getall")

	data := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, data)
}

func (p *EmployeeGroupService) AddGroupToCategory(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID   int64  `json:"category_id" binding:"required,gt=0"`
		TypeActionID string `json:"typeaction_id" binding:"required,min=5,max=50"`
		GroupID      int64  `json:"group_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.CategoryID, body.TypeActionID, body.GroupID, employeeID)
	pg.SetStoredProcedureName("tickets.categorygroup_add")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) UpdateGroupForCategory(ctx *gin.Context, params map[string]interface{}) {

	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID   int64  `json:"category_id" binding:"required,gt=0"`
		TypeActionID string `json:"typeaction_id" binding:"required"`
		GroupID      int64  `json:"group_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.CategoryID, body.TypeActionID, body.GroupID, employeeID)
	pg.SetStoredProcedureName("tickets.categorygroup_upd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) DeleteGroupToCategory(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CategoryID   int64  `json:"category_id" binding:"required,gt=0"`
		TypeActionID string `json:"typeaction_id" binding:"required,min=5,max=50"`
		GroupID      int64  `json:"group_id" binding:"required,LteInt32"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.CategoryID, body.TypeActionID, body.GroupID, employeeID)
	pg.SetStoredProcedureName("tickets.categorygroup_delete")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (p *EmployeeGroupService) UpdateGroupName(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		GroupID   int64  `json:"group_id" binding:"required,LteInt32"`
		GroupName string `json:"group_name" binding:"required,max=100"`
	})
	if !ok {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		p.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(body.GroupID, body.GroupName, employeeID)
	pg.SetStoredProcedureName("hr.group_groupnameupd")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
