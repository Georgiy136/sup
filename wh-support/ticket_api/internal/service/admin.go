package service

import (
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/service/chat_mapper"
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
)

type AdminService struct {
	validator                    *validator.Validate
	errBuilder                   core_errors.ErrorBuilder
	redisClient                  RedisClientInterface
	resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface
	chatMapper                   *chat_mapper.ChatMapperStreaming
}

func NewAdminService(redisClient RedisClientInterface, resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface, chatMapper *chat_mapper.ChatMapperStreaming) *AdminService {
	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)

	return &AdminService{
		validator:                    valid,
		errBuilder:                   core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
		redisClient:                  redisClient,
		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
		chatMapper:                   chatMapper,
	}
}

func (ad *AdminService) GetTicketsByCategoryIDs(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction  string  `json:"type_action" binding:"required"`
		CategoryIDs []int64 `json:"category_ids" binding:"required,gt=0"`
	})
	if !ok {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	accessPolicyResult, err := ad.redisClient.GetAccessPolicy(ctx, employeeID, body.TypeAction)
	if err != nil {
		ad.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get resources from cache: %w", err))
		return
	}

	externalActionsByEmployeeID, err := ad.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPolicyResult.EmployeeGroups) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	categoriesEmployee := make(map[int64]struct{})
	access.CollectAccessibleCategoriesByGroups(categoriesEmployee, accessPolicyResult.AccessData.ActionGroups, accessPolicyResult.EmployeeGroups)
	access.CollectAccessibleCategoriesByExternalActions(categoriesEmployee, accessPolicyResult.AccessData.ExternalActions, externalActionsByEmployeeID)

	filteredCategories := make([]int64, 0)
	for _, category := range body.CategoryIDs {
		if _, ok := categoriesEmployee[category]; ok {
			filteredCategories = append(filteredCategories, category)
		}
	}

	if len(filteredCategories) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(filteredCategories)
	pg.SetStoredProcedureName("tickets.tickets_getallbycategories")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (ad *AdminService) GetCenterTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction  string  `json:"type_action" binding:"required"`
		CategoryIDs []int64 `json:"category_ids" binding:"required,min=1,max=5"`
	})
	if !ok {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	accessPolicyResult, err := ad.redisClient.GetAccessPolicy(ctx, employeeID, body.TypeAction)
	if err != nil {
		ad.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get resources from cache: %w", err))
		return
	}

	externalActionsByEmployeeID, err := ad.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPolicyResult.EmployeeGroups) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	categoriesEmployee := make(map[int64]struct{})
	access.CollectAccessibleCategoriesByGroups(categoriesEmployee, accessPolicyResult.AccessData.ActionGroups, accessPolicyResult.EmployeeGroups)
	access.CollectAccessibleCategoriesByExternalActions(categoriesEmployee, accessPolicyResult.AccessData.ExternalActions, externalActionsByEmployeeID)

	filteredCategories := make([]int64, 0)
	for _, category := range body.CategoryIDs {
		if _, ok := categoriesEmployee[category]; ok {
			filteredCategories = append(filteredCategories, category)
		}
	}

	if len(filteredCategories) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(filteredCategories)
	pg.SetStoredProcedureName("tickets.tickets_getallforcenter")

	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := ad.chatMapper.MapTicketListWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetCenterTickets] MapTicketListWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (ad *AdminService) GetViewTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	accessPolicyResult, err := ad.redisClient.GetAccessPolicy(ctx, employeeID, "view_ticket_category")
	if err != nil {
		ad.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get resources from cache: %w", err))
		return
	}

	externalActionsByEmployeeID, err := ad.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPolicyResult.EmployeeGroups) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	accessibleCategories := access.DetermineAccessibleCategories(
		accessPolicyResult.EmployeeGroups,
		externalActionsByEmployeeID,
		accessPolicyResult.AccessData,
	)

	if len(accessibleCategories) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(accessibleCategories)
	pg.SetStoredProcedureName("tickets.tickets_getallforcenter")

	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}
	res, err := ad.chatMapper.MapTicketListWithChatInfo(ctx, rd.GetData(), employeeID)
	if err != nil {
		logrus.Errorf("[GetViewTickets] MapTicketListWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func (ad *AdminService) GetCenterTicketIDs(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	accessibleCategories, errKey, err := ad.getCategoriesForViewTicketByEmployeeID(employeeID, ctx)
	if err != nil {
		ad.errBuilder.BindError(ctx, errKey, err)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(accessibleCategories)
	pg.SetStoredProcedureName("tickets.tickets_getticketidsforcenter")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (ad *AdminService) GetCenterTicketByTicketIDs(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		ad.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketIDs []int64 `json:"ticket_ids" binding:"required,gt=0,lte=100,dive,gt=0"`
	})
	if !ok {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		ad.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	accessibleCategories, errKey, err := ad.getCategoriesForViewTicketByEmployeeID(employeeID, ctx)
	if err != nil {
		ad.errBuilder.BindError(ctx, errKey, err)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketIDs, accessibleCategories)
	pg.SetStoredProcedureName("tickets.tickets_getallforcenterbyticketids")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (ad *AdminService) getCategoriesForViewTicketByEmployeeID(employeeID int64, ctx *gin.Context) ([]int64, core_errors.ErrorKey, error) {
	viewTypeAction := "view_ticket_category"

	accessPolicyResult, err := ad.redisClient.GetAccessPolicy(ctx, employeeID, viewTypeAction)
	if err != nil {
		return nil, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get resources from cache: %w", err)
	}

	externalActionsByEmployeeID, err := ad.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		return nil, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err)
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPolicyResult.EmployeeGroups) == 0 {
		return nil, "", nil
	}

	return access.DetermineAccessibleCategories(
		accessPolicyResult.EmployeeGroups,
		externalActionsByEmployeeID,
		accessPolicyResult.AccessData,
	), "", nil
}
