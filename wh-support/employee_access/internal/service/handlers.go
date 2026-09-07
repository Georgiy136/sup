package service

import (
	"context"
	"errors"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_access/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_access/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

type EmployeeAccessService struct {
	redisClient *clients.RedisClient
	errBuilder  core_errors.ErrorBuilder
}

func NewEmployeeAccessService() *EmployeeAccessService {
	return &EmployeeAccessService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (r *EmployeeAccessService) Init(ctx context.Context, cfg configs.Config) {
	r.redisClient = clients.NewRedisClient(ctx, cfg)
}

func (r *EmployeeAccessService) GetEmployeeAllowedActionsHandler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		Actions []string `json:"actions" binding:"required"`
	})
	if !ok {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		r.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)

	cacheActions, err := r.redisClient.GetFromCache(ctx, employeeID)
	if err != nil {
		r.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, errors.New("can't get resources from cache"))
		return
	}

	if cacheActions == nil {
		utils.BindNoContent(ctx)
		return
	}

	cacheActionsMap := make(map[string]struct{}, len(cacheActions.ActionIds))
	for _, action := range cacheActions.ActionIds {
		cacheActionsMap[action] = struct{}{}
	}

	allowedActions := []models.AppActions{}
	for _, action := range body.Actions {
		if _, ok = cacheActionsMap[action]; ok {
			allowedActions = append(allowedActions, models.AppActions{Action: action})
		}
	}

	if len(allowedActions) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, models.ListEmployeeResources{AppActions: allowedActions})
}
