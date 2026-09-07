package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/sirupsen/logrus"
)

func (h *BandActionHandler) HandleBandAdditionalInfoDialog(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CallbackID string          `json:"callback_id" binding:"required,lte=30"`
		State      string          `json:"state" binding:"required,lte=1000"`
		UserID     string          `json:"user_id" binding:"required,lte=30"`
		ChannelID  string          `json:"channel_id" binding:"required,lte=30"`
		Submission json.RawMessage `json:"submission" binding:"required"`
		Cancelled  bool            `json:"cancelled"`
	})
	if !ok {
		utils.BindValidationErrorWithAbort(ctx, "unsupported band additional info dialog body")
		return
	}
	if body == nil {
		utils.BindValidationErrorWithAbort(ctx, "can't parse band additional info dialog body")
		return
	}

	var state models.BandDialogState
	if err := jsoniter.Unmarshal([]byte(body.State), &state); err != nil {
		utils.BindValidationErrorWithAbort(ctx, fmt.Sprintf("parse additional info dialog state error: %v", err))
		return
	}

	if !h.isAuthorized(models.BandActionContext{
		TicketID:         state.TicketID,
		Action:           state.Action,
		Operation:        state.Operation,
		TypeOfEmployeeID: state.TypeOfEmployeeID,
		EmployeeID:       state.EmployeeID,
		Signature:        state.Signature,
		Timestamp:        state.Timestamp,
	}) {
		logrus.Warnf("band additional info dialog callback rejected: unauthorized user_agent=%s", ctx.Request.UserAgent())
		utils.BindAuthenticationErrorWithAbort(ctx, "band.additional_info.unauthorized", "Запрос не авторизован", "invalid band action token")
		return
	}

	if body.Cancelled {
		ctx.JSON(http.StatusOK, model.SubmitDialogResponse{})
		utils.BindNoContent(ctx)
		return
	}

	if body.CallbackID != consts.BandAdditionalInfoDialogCallbackID {
		utils.BindValidationErrorWithAbort(ctx, fmt.Sprintf("unexpected callback id: %s", body.CallbackID))
		return
	}

	var submission map[string]any
	if err := jsoniter.Unmarshal(body.Submission, &submission); err != nil {
		utils.BindServiceErrorWithAbort(ctx, "unmarshal submission failed", err)
		return
	}
	actionRequest := models.BandActionRequest{
		UserID:           body.UserID,
		EmployeeID:       state.EmployeeID,
		ChannelID:        body.ChannelID,
		PostID:           state.PostID,
		TicketID:         state.TicketID,
		Action:           state.Action,
		Operation:        state.Operation,
		TypeOfEmployeeID: state.TypeOfEmployeeID,
	}

	if err := h.bandActionService.HandleBandAdditionalInfoDialog(ctx.Request.Context(), actionRequest, submission, state); err != nil {
		if validationErr, ok := errors.AsType[*customerrors.InteractiveDialogValidationError](err); ok {
			ctx.JSON(http.StatusOK, model.SubmitDialogResponse{Errors: validationErr.Errors})
			utils.BindNoContent(ctx)
			return
		}
		utils.BindServiceErrorWithAbort(ctx, fmt.Sprintf("can't handle additional info dialog, action %s for ticket %d, category_id %d, status_id %s: %v", state.Action, state.TicketID, state.CategoryID, state.StatusID, err), err)
		return
	}

	ctx.JSON(http.StatusOK, model.SubmitDialogResponse{})
	utils.BindNoContent(ctx)
}
