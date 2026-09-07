package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/consts"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"
)

type CategoryInfo struct {
	validator                    *validator.Validate
	errBuilder                   core_errors.ErrorBuilder
	accessPolicyCache            AccessPolicyCache
	resourceEmployeeAccessClient ResourceEmployeeAccessClient
	categoryAccessResolver       *access.CategoryAccessResolver
}

func NewCategoryInfo(accessPolicyCache AccessPolicyCache, resourceEmployeeAccessClient ResourceEmployeeAccessClient) *CategoryInfo {
	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)

	return &CategoryInfo{
		validator:                    valid,
		errBuilder:                   core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
		accessPolicyCache:            accessPolicyCache,
		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
		categoryAccessResolver:       access.NewCategoryAccessResolver(),
	}
}

func (c *CategoryInfo) GetCategoriesTree(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TypeAction string `json:"type_action" binding:"required"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type action"))
		return
	}

	accessData, err := c.accessPolicyCache.GetAccessData(ctx, body.TypeAction, employeeID)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get access data from redis: %w", err))
		return
	}

	externalActionsByEmployeeID, err := c.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessData.EmployeeGroups) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	accessibleCategories := c.categoryAccessResolver.DetermineAccessibleCategories(
		accessData.ExternalActions,
		accessData.ActionGroups,
		externalActionsByEmployeeID,
		accessData.EmployeeGroups,
	)

	if len(accessibleCategories) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(consts.SupportPgDatabaseKey)
	pg.SetParams(accessibleCategories)
	pg.SetStoredProcedureName("tickets.category_getbycategories")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (c *CategoryInfo) GetCategoriesTreeForCreateTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	accessData, err := c.accessPolicyCache.GetAccessData(ctx, "create_ticket_category", employeeID)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get access data from redis: %w", err))
		return
	}

	externalActionsByEmployeeID, err := c.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	accessibleCategories := c.categoryAccessResolver.DetermineAccessibleCategories(
		accessData.ExternalActions,
		accessData.ActionGroups,
		externalActionsByEmployeeID,
		accessData.EmployeeGroups,
	)

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(consts.SupportPgDatabaseKey)
	pg.SetParams(accessibleCategories)
	pg.SetStoredProcedureName("tickets.category_getbycategoriesforcreate")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
