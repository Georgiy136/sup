package service

import (
	"context"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/statistics_api/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/statistics_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

type RedisClientInterface interface {
	GetAccessPolicies(ctx context.Context, employeeID int64, typeActions ...string) (*models.AccessPoliciesResult, error)
}

type ResourceEmployeeAccessClientInterface interface {
	GetAccessActionsByEmployeeID(employeeID int64) ([]string, error)
}

type Statistics struct {
	validator                    *validator.Validate
	errBuilder                   core_errors.ErrorBuilder
	redisClient                  RedisClientInterface
	resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface
}

func NewStatistics(redisClient RedisClientInterface, resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface) *Statistics {
	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)

	return &Statistics{
		validator:                    valid,
		errBuilder:                   core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
		redisClient:                  redisClient,
		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
	}
}

func (s *Statistics) GetStatisticsByEmployeeV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.ticketsstatistics_getbyemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}

func (s *Statistics) GetTicketStatistics(ctx *gin.Context, params map[string]interface{}) {
	const (
		workCategoriesApproveActionType = "status_approve_category"
		workCategoriesPerformActionType = "status_perform_category"
	)

	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	accessPoliciesResult, err := s.redisClient.GetAccessPolicies(ctx, employeeID, workCategoriesApproveActionType, workCategoriesPerformActionType)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get access policies from cache: %w", err))
		return
	}

	approveAccessData := accessPoliciesResult.Policies[workCategoriesApproveActionType]
	performAccessData := accessPoliciesResult.Policies[workCategoriesPerformActionType]

	externalActionsByEmployeeID, err := s.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all access actions by employee ID: %w", err))
		return
	}

	categoryStatusAccess := access.DetermineCategoryStatusAccess(
		accessPoliciesResult.EmployeeGroups,
		externalActionsByEmployeeID,
		approveAccessData,
		performAccessData,
	)

	if len(categoryStatusAccess) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	categoryStatusIDsJSON, err := jsoniter.Marshal(categoryStatusAccess)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal category status access: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(string(categoryStatusIDsJSON), employeeID)
	pg.SetStoredProcedureName("tickets.ticketsstatistics_getbyemployee_v2")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
