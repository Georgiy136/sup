package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/models"
	internalutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/utils"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

const (
	workCategoriesApproveActionType = "status_approve_category"
	workCategoriesPerformActionType = "status_perform_category"
)

type ChatAuthService struct {
	accessTokenTTL time.Duration

	refreshWindow                time.Duration
	jwtGenerator                 jwtTokenGeneratorInterface
	redisClient                  redisClientInterface
	resourceEmployeeAccessClient resourceEmployeeAccessClientInterface
	errBuilder                   core_errors.ErrorBuilder
}

func NewChatAuthService(jwtGenerator jwtTokenGeneratorInterface, redisClient redisClientInterface, resourceEmployeeAccessClient resourceEmployeeAccessClientInterface) *ChatAuthService {
	return &ChatAuthService{
		errBuilder:   core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
		jwtGenerator: jwtGenerator,
		redisClient:  redisClient,

		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
	}
}

func (c *ChatAuthService) Configure(ctx context.Context, config configs.Config) {
	var params models.AuthTokenParams
	err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("auth_token_params"), &params)
	if err != nil {
		logrus.Panicf("cannot unmarshal auth_token_params: %v", err)
	}
	if c.accessTokenTTL, err = time.ParseDuration(params.AuthTokenTTLDuration); err != nil {
		logrus.Panicf("cannot parse %s as duration: %v", params.AuthTokenTTLDuration, err)
	}
	if c.refreshWindow, err = time.ParseDuration(params.RefreshWindow); err != nil {
		logrus.Panicf("cannot parse %s as duration: %v", params.RefreshWindow, err)
	}
}

func (c *ChatAuthService) GetRealtimeToken(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't parse employee_id parameter"))
		return
	}

	token, err := c.jwtGenerator.GenerateToken(employeeID, c.accessTokenTTL)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("failed to generate token: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, models.GenerateTokenResponse{Token: token})
}

func (c *ChatAuthService) ChatAuthSubscribe(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		Channel string `json:"channel" binding:"required,lte=255"`
		User    string `json:"user" binding:"required"`
	})
	if !ok {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, "unsupported body type")
		return
	}
	if body == nil {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, "empty body")
		return
	}

	// Парсим employee_id из поля user (Centrifugo передаёт sub из JWT токена)
	employeeID, err := strconv.ParseInt(body.User, 10, 64)
	if err != nil {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid user id: %v", err))
		return
	}

	// Извлекаем ticket_id из названия канала
	ticketID, err := internalutils.ExtractTicketIDFromChannel(body.Channel)
	if err != nil {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid channel format: %v", err))
		return
	}

	// Проверяем доступ к заявке
	hasAccess, err := c.checkTicketAccess(ctx, employeeID, ticketID)
	if err != nil {
		internalutils.BindCentrifugoError(ctx, http.StatusInternalServerError, fmt.Sprintf("internal server error: %v", err))
		return
	}
	if !hasAccess {
		logrus.WithFields(logrus.Fields{
			"employee_id": employeeID,
			"ticket_id":   ticketID,
			"channel":     body.Channel,
			"user":        body.User,
		}).Info("chat auth subscribe permission denied")
		internalutils.BindCentrifugoError(ctx, http.StatusForbidden, "permission denied")
		return
	}

	internalutils.BindCentrifugoSuccess(ctx, nil)
}

func (c *ChatAuthService) ChatAuthRefresh(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		User string `json:"user" binding:"required"`
	})
	if !ok {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, "unsupported body type")
		return
	}
	if body == nil {
		internalutils.BindCentrifugoError(ctx, http.StatusBadRequest, "empty body")
		return
	}

	expireAt := time.Now().Add(c.accessTokenTTL).Unix()

	internalutils.BindCentrifugoSuccess(ctx, gin.H{
		"expired":   false,
		"expire_at": expireAt,
	})
}

func (c *ChatAuthService) checkTicketAccess(ctx context.Context, employeeID int64, ticketID int64) (bool, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(ticketID)
	pg.SetStoredProcedureName("tickets.tickets_getforchat")

	ticketInfo := models.TicketAccessInfo{}
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &ticketInfo)
	if rd.HasError() {
		return false, fmt.Errorf("get ticket info error: %w", rest_data.ToError(rd))
	}

	baseLogFields := logrus.Fields{
		"employee_id":                  employeeID,
		"ticket_id":                    ticketID,
		"category_id":                  ticketInfo.CategoryID,
		"create_employee_id":           ticketInfo.CreateEmployeeID,
		"group_employee_ids_count":     len(ticketInfo.GroupEmployeeIDs),
		"favourite_employee_ids_count": len(ticketInfo.FavouriteEmployeeIDs),
	}
	ticketEmployeeIDs := slices.Concat(ticketInfo.GroupEmployeeIDs, ticketInfo.FavouriteEmployeeIDs)
	if ticketInfo.CreateEmployeeID == employeeID || slices.Contains(ticketEmployeeIDs, employeeID) {
		logrus.WithFields(baseLogFields).Info("ticket chat access granted by direct ticket relation")
		return true, nil
	}

	// Проверяем внутренние и внешние доступы сотрудника к заявке
	accessPoliciesResult, err := c.redisClient.GetAccessPolicies(ctx, employeeID, workCategoriesApproveActionType, workCategoriesPerformActionType)
	if err != nil {
		return false, fmt.Errorf("can't get access policies from cache: %w", err)
	}

	approveAccessData := accessPoliciesResult.Policies[workCategoriesApproveActionType]
	performAccessData := accessPoliciesResult.Policies[workCategoriesPerformActionType]

	externalActionsByEmployeeID, err := c.resourceEmployeeAccessClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		return false, fmt.Errorf("error get all access actions by employee ID: %w", err)
	}

	accessLogFields := logrus.Fields{
		"employee_id":                     employeeID,
		"ticket_id":                       ticketID,
		"category_id":                     ticketInfo.CategoryID,
		"create_employee_id":              ticketInfo.CreateEmployeeID,
		"employee_groups":                 accessPoliciesResult.EmployeeGroups,
		"external_actions_by_employee_id": externalActionsByEmployeeID,
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPoliciesResult.EmployeeGroups) == 0 {
		logrus.WithFields(accessLogFields).Info("ticket chat access denied: employee has no groups or external actions")

		return false, nil
	}

	accessibleCategories := access.DetermineAccessibleCategories(
		accessPoliciesResult.EmployeeGroups,
		externalActionsByEmployeeID,
		approveAccessData,
		performAccessData,
	)

	accessLogFields["accessible_categories"] = accessibleCategories

	if len(accessibleCategories) == 0 {
		logrus.WithFields(accessLogFields).Info("ticket chat access denied: accessible categories are empty")
		return false, nil
	}

	if !slices.Contains(accessibleCategories, ticketInfo.CategoryID) {
		logrus.WithFields(accessLogFields).Info("ticket chat access denied: ticket category is not accessible")
		return false, nil
	}

	logrus.WithFields(accessLogFields).Info("ticket chat access granted by category access")
	return true, nil
}
