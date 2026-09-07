package handlers

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	bot       bot
	validator *validator.Validate
}

func New(b bot, valid *validator.Validate) *Handlers {
	return &Handlers{
		bot:       b,
		validator: valid,
	}
}

func (h *Handlers) Configure(_ context.Context, _ configs.Config) {
	gocore_validators.InitializeCustomValidatorsV10(h.validator)
}

func (h *Handlers) SendNotificationHandler(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.RawBODY].([]byte)
	if !ok {
		utils.BindValidationErrorWithAbort(ctx, "can't parse body")
		return
	}
	if len(body) == 0 {
		utils.BindValidationErrorWithAbort(ctx, "passed empty body")
		return
	}

	var bodyStruct models.Notification
	if err := jsoniter.Unmarshal(body, &bodyStruct); err != nil {
		utils.BindValidationErrorWithAbort(ctx, "cannot unmarshal body: "+err.Error())
		return
	}
	if err := h.validator.Struct(bodyStruct); err != nil {
		utils.BindValidationErrorWithAbort(ctx, "validation: "+err.Error())
		return
	}

	err := h.bot.SendNotification(bodyStruct)
	if err != nil {
		logrus.Errorf("can't send notification: %v", err)
	}

	utils.BindNoContent(ctx)
}

type bot interface {
	SendNotification(notifications models.Notification) error
}
