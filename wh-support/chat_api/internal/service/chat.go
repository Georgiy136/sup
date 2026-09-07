package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/repository"
	internalutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/utils"

	employeetags "gitlab.wildberries.ru/wbwh/support/utils.git/employee_tags"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type ChatService struct {
	errBuilder                      core_errors.ErrorBuilder
	repo                            ticketsRepo
	chatCache                       chatCache
	chatProviderApiClient           chatProviderApiClient
	resourceEmployeeAccessApiClient resourceEmployeeAccessApiClient
	employeeInfoApiClient           employeeInfoApiClient
}

func NewChatService(
	repo ticketsRepo,
	chatCache chatCache,
	chatProviderApiClient chatProviderApiClient,
	resourceEmployeeAccessApiClient resourceEmployeeAccessApiClient,
	employeeInfoApiClient employeeInfoApiClient) *ChatService {

	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)

	errorsRepo := errors_keys.NewErrorsRepository()
	errorsRepo.SetErrors(support_err_keys.ErrorsRepository)

	return &ChatService{
		errBuilder:                      core_errors.NewErrorBuilder(errorsRepo.GetErrors(), nil),
		repo:                            repo,
		chatCache:                       chatCache,
		chatProviderApiClient:           chatProviderApiClient,
		resourceEmployeeAccessApiClient: resourceEmployeeAccessApiClient,
		employeeInfoApiClient:           employeeInfoApiClient,
	}
}

func (c *ChatService) GetChat(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, chatID, err := c.checkTicketAccessAndGetChatID(ctx, employeeID, body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("cant't check employee access to ticket: %w", err))
		return
	}
	if !hasAccess {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for get chat"))
		return
	}

	if chatID != "" {
		chat, err := c.chatProviderApiClient.GetChatByChatID(ctx, employeeID, models.GetChatByChatIDRequest{ChatID: chatID})
		if err != nil {
			var dataError clients.DataError
			if errors.As(err, &dataError) {
				c.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("can't get chat by id: %w", dataError))
				return
			}

			c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get chat by id: %w", err))
			return
		}

		utils.BindObjectToRestData(ctx, chat)
		return
	}

	newChat, err := c.chatProviderApiClient.CreateNewChat(ctx, employeeID, models.CreateNewChatRequest{TicketID: body.TicketID})
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't create new chat: %w", err))
		return
	}

	err = c.repo.AddBandChat(newChat.Chat.ChatID, newChat.TicketID, newChat.Chat.EmployeeID)
	if err != nil {
		if errBiz, ok := errors.AsType[*repository.DBBizError](err); ok {
			utils.BindRestData(ctx, rest_data.RestData{
				Errors: []rest_data.CustomError{{
					ErrorKey: errBiz.ErrorKey,
					Message:  errBiz.Message,
					Detail:   errBiz.Detail,
				}},
				HttpResultCode: http.StatusUnprocessableEntity,
			})
			return
		}

		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't add create connection between chat and ticket: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, newChat)
}

func (c *ChatService) CreatePost(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64    `json:"ticket_id" binding:"required,gt=0"`
		Message  string   `json:"message" binding:"required,lte=5000"`
		FileIDs  []string `json:"file_ids" binding:"omitempty,dive,lte=30"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, chatID, err := c.checkTicketAccessAndGetChatID(ctx, employeeID, body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("cant't check employee access to ticket: %w", err))
		return
	}
	if !hasAccess {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for create post"))
		return
	}

	if chatID == "" {
		c.errBuilder.BindError(ctx, support_err_keys.KeyErrorChatNotFound, fmt.Errorf("chat doesn't exist"))
		return
	}

	if employeeIDs := employeetags.GetTaggedEmployeeIDs(body.Message); len(employeeIDs) > 0 {
		ticketEmployeesForTag, err := c.repo.GetTicketEmployeeIDsForTag(body.TicketID)
		if err != nil {
			c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get ticket employees for tag in chat: %w", err))
			return
		}

		if ticketEmployeesForTag == nil || len(ticketEmployeesForTag.EmployeeIDs) == 0 {
			c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("no employees available for tagging in this chat"))
			return
		}

		employeeIDsForTagMap := internalutils.SliceToMap(ticketEmployeesForTag.EmployeeIDs)
		for _, employeeID := range employeeIDs {
			if _, exist := employeeIDsForTagMap[employeeID]; !exist {
				c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("employee cannot be tagged in chat"))
				return
			}
		}
	}

	post, err := c.chatProviderApiClient.CreatePost(ctx, employeeID, models.CreatePostRequest{
		ChatID:     chatID,
		MessageStr: body.Message,
		TicketID:   body.TicketID,
		FileIDs:    body.FileIDs,
	})
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't create post: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, post)
}

func (c *ChatService) GetHistoryChat(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID     int64  `json:"ticket_id" binding:"required,gt=0"`
		CountPost    int64  `json:"count_post" binding:"required,gt=0,lte=200"`
		FromPost     string `json:"from_post" binding:"required,lte=30"`
		FromCreateAt string `json:"from_create_at" binding:"required,IsRFC3339Nano"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, chatID, err := c.checkTicketAccessAndGetChatID(ctx, employeeID, body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("cant't check employee access to ticket: %w", err))
		return
	}
	if !hasAccess {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for get history chat"))
		return
	}

	if chatID == "" {
		c.errBuilder.BindError(ctx, support_err_keys.KeyErrorChatNotFound, fmt.Errorf("chat doesn't exist"))
		return
	}

	historyChat, err := c.chatProviderApiClient.GetHistoryChatByChatID(ctx, employeeID, models.GetHistoryChatByChatIDRequest{
		ChatID:       chatID,
		FromPost:     body.FromPost,
		CountPost:    body.CountPost,
		FromCreateAt: body.FromCreateAt,
	})
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get history chat: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, historyChat)
}

func (c *ChatService) ViewChatByUser(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, chatID, err := c.checkTicketAccessAndGetChatID(ctx, employeeID, body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("cant't check employee access to ticket: %w", err))
		return
	}
	if !hasAccess {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for view chat"))
		return
	}

	if chatID == "" {
		c.errBuilder.BindError(ctx, support_err_keys.KeyErrorChatNotFound, fmt.Errorf("chat doesn't exist"))
		return
	}

	now := time.Now().Format(time.RFC3339Nano)
	err = c.chatCache.AddLastChatViewByUser(ctx, employeeID, body.TicketID, now)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't add last view chat by user: %w", err))
		return
	}

	bodyResponse := models.ViewChatByUserResponse{
		TicketID:     body.TicketID,
		TimeViewChat: now,
	}

	utils.BindObjectToRestData(ctx, bodyResponse)
}

func (c *ChatService) GetTicketEmployeesForChatSelector(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, chatID, err := c.checkTicketAccessAndGetChatID(ctx, employeeID, body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("cant't check employee access to ticket: %w", err))
		return
	}
	if !hasAccess {
		c.errBuilder.BindError(ctx, errors_keys.ErrVldTokenAccessWrong, fmt.Errorf("no employee access for view chat"))
		return
	}

	if chatID == "" {
		c.errBuilder.BindError(ctx, support_err_keys.KeyErrorChatNotFound, fmt.Errorf("chat doesn't exist"))
		return
	}

	ticketEmployeesForTag, err := c.repo.GetTicketEmployeeIDsForTag(body.TicketID)
	if err != nil {
		c.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, fmt.Errorf("can't get ticket employee ids for tag in chat: %w", err))
		return
	}

	if ticketEmployeesForTag == nil || len(ticketEmployeesForTag.EmployeeIDs) == 0 {
		utils.BindNoContent(ctx)
	}

	employeesInfo, err := c.employeeInfoApiClient.GetEmployeesFullName(employeeID, ticketEmployeesForTag.EmployeeIDs)
	if err != nil {
		c.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get employees info for tag in chat: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, employeesInfo)
}

func (c *ChatService) checkTicketAccessAndGetChatID(ctx context.Context, employeeID int64, ticketID int64) (bool, string, error) {
	ticketInfo, err := c.repo.GetTicketInfoByTicketID(ticketID)
	if err != nil {
		return false, "", fmt.Errorf("can't get ticket info from db: %w", err)
	}

	if ticketInfo == nil {
		return false, "", nil
	}

	chatID := ""
	if ticketInfo.ChatID != nil {
		chatID = *ticketInfo.ChatID
	}

	accessPoliciesResult, err := c.chatCache.GetAccessPolicies(ctx, employeeID, models.WorkCategoriesApproveActionType, models.WorkCategoriesPerformActionType, models.WorkCategoriesViewActionType)
	if err != nil {
		return false, "", fmt.Errorf("can't get access policies from cache: %w", err)
	}

	externalActionsByEmployeeID, err := c.resourceEmployeeAccessApiClient.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		return false, "", fmt.Errorf("error get all access actions by employee ID: %w", err)
	}

	approveAccessData := accessPoliciesResult.Policies[models.WorkCategoriesApproveActionType]
	performAccessData := accessPoliciesResult.Policies[models.WorkCategoriesPerformActionType]
	viewAccessData := accessPoliciesResult.Policies[models.WorkCategoriesViewActionType]

	accessibleCategories := access.DetermineAccessibleCategories(
		accessPoliciesResult.EmployeeGroups,
		externalActionsByEmployeeID,
		approveAccessData,
		performAccessData,
		viewAccessData,
	)

	accessLogFields := logrus.Fields{
		"employee_id":           employeeID,
		"ticket_id":             ticketID,
		"category_id":           ticketInfo.CategoryID,
		"create_employee_id":    ticketInfo.CreateEmployeeID,
		"employee_groups":       accessPoliciesResult.EmployeeGroups,
		"approve_access_data":   approveAccessData,
		"perform_access_data":   performAccessData,
		"view_access_data":      viewAccessData,
		"accessible_categories": accessibleCategories,
	}
	logrus.WithFields(accessLogFields).Info("access info from cache")

	baseLogFields := logrus.Fields{
		"employee_id":            employeeID,
		"ticket_id":              ticketID,
		"chat_id":                chatID,
		"category_id":            ticketInfo.CategoryID,
		"create_employee_id":     ticketInfo.CreateEmployeeID,
		"group_employee_ids":     ticketInfo.GroupEmployeeIDs,
		"favourite_employee_ids": ticketInfo.FavouriteEmployeeIDs,
	}
	logrus.WithFields(baseLogFields).Info("ticket info from db")

	ticketEmployeeIDs := slices.Concat(ticketInfo.GroupEmployeeIDs, ticketInfo.FavouriteEmployeeIDs)
	if ticketInfo.CreateEmployeeID == employeeID || slices.Contains(ticketEmployeeIDs, employeeID) {
		return true, chatID, nil
	}

	return false, "", nil
}
