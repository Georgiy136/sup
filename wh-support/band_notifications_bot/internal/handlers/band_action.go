package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *BandActionHandler) HandleBandAction(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UserID    string          `json:"user_id" binding:"required,lte=30"`
		ChannelID string          `json:"channel_id" binding:"required,lte=30"`
		PostID    string          `json:"post_id" binding:"required,lte=30"`
		TriggerID string          `json:"trigger_id" binding:"omitempty,lte=500"`
		Context   json.RawMessage `json:"context" binding:"required"`
	})
	if !ok {
		utils.BindValidationErrorWithAbort(ctx, "unsupported band action body")
		return
	}
	if body == nil {
		utils.BindValidationErrorWithAbort(ctx, "can't parse band action body")
		return
	}
	if body.Context == nil {
		utils.BindValidationErrorWithAbort(ctx, "body context is nil")
		return
	}

	var actionContext models.BandActionContext
	if err := jsoniter.Unmarshal(body.Context, &actionContext); err != nil {
		utils.BindValidationErrorWithAbort(ctx, "unmarshal request context failed")
		return
	}
	if !h.isAuthorized(actionContext) {
		logrus.Warnf("band action callback rejected: unauthorized user_agent=%s", ctx.Request.UserAgent())
		utils.BindAuthenticationErrorWithAbort(ctx, "band.action.unauthorized", "Запрос не авторизован", "invalid band action token")
		return
	}

	actionRequest := models.BandActionRequest{
		UserID:           body.UserID,
		EmployeeID:       actionContext.EmployeeID,
		ChannelID:        body.ChannelID,
		PostID:           body.PostID,
		TriggerID:        body.TriggerID,
		TicketID:         actionContext.TicketID,
		CategoryID:       actionContext.CategoryID,
		Action:           actionContext.Action,
		Operation:        actionContext.Operation,
		TypeOfEmployeeID: actionContext.TypeOfEmployeeID,
		StatusID:         actionContext.StatusID,
	}

	if err := h.bandActionService.HandleBandAction(ctx.Request.Context(), actionRequest); err != nil {
		utils.BindServiceErrorWithAbort(ctx, fmt.Sprintf("can't handle band action %s for ticket %d, category_id %d, status_id %s: %v", actionContext.Action, actionRequest.TicketID, actionRequest.CategoryID, actionRequest.StatusID, err), err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
	utils.BindNoContent(ctx)
}
