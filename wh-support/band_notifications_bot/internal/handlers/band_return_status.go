package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

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

func (h *BandActionHandler) HandleBandReturnToStatusDialog(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		CallbackID string          `json:"callback_id" binding:"required,lte=30"`
		State      string          `json:"state" binding:"required,lte=1000"`
		UserID     string          `json:"user_id" binding:"required,lte=30"`
		ChannelID  string          `json:"channel_id" binding:"required,lte=30"`
		Submission json.RawMessage `json:"submission" binding:"required"`
		Cancelled  bool            `json:"cancelled"`
	})
	if !ok {
		utils.BindValidationErrorWithAbort(ctx, "unsupported band return to status dialog body")
		return
	}
	if body == nil {
		utils.BindValidationErrorWithAbort(ctx, "can't parse band return to status dialog body")
		return
	}

	var state models.BandDialogState
	if err := jsoniter.Unmarshal([]byte(body.State), &state); err != nil {
		utils.BindValidationErrorWithAbort(ctx, fmt.Sprintf("parse return to status dialog state error: %v", err))
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
		logrus.Warnf("band return dialog callback rejected: unauthorized user_agent=%s", ctx.Request.UserAgent())
		utils.BindAuthenticationErrorWithAbort(ctx, "band.return.unauthorized", "Запрос не авторизован", "invalid band action token")
		return
	}

	if body.Cancelled {
		ctx.JSON(http.StatusOK, model.SubmitDialogResponse{})
		utils.BindNoContent(ctx)
		return
	}

	actionRequest, err := convertReturnToStatusDialogRequest(models.BandSubmitDialogRequest{
		CallbackID: body.CallbackID,
		State:      body.State,
		UserID:     body.UserID,
		ChannelID:  body.ChannelID,
		Submission: body.Submission,
		Cancelled:  body.Cancelled,
	}, state)
	if err != nil {
		if validationErr, ok := errors.AsType[*customerrors.InteractiveDialogValidationError](err); ok {
			ctx.JSON(http.StatusOK, model.SubmitDialogResponse{Errors: validationErr.Errors})
			utils.BindNoContent(ctx)
			return
		}
		utils.BindServiceErrorWithAbort(ctx, fmt.Sprintf("convert return to status dialog request error: %v", err), err)
		return
	}

	if err = h.bandActionService.HandleBandReturnToStatusDialog(ctx.Request.Context(), *actionRequest); err != nil {
		utils.BindServiceErrorWithAbort(ctx, fmt.Sprintf("can't handle band return to status dialog, action %s for ticket %d, category_id %d, status_id %s: %v", state.Action, state.TicketID, state.CategoryID, state.StatusID, err), err)
		return
	}

	ctx.JSON(http.StatusOK, model.SubmitDialogResponse{})
	utils.BindNoContent(ctx)
}

func convertReturnToStatusDialogRequest(request models.BandSubmitDialogRequest, state models.BandDialogState) (*models.BandActionRequest, error) {
	if request.CallbackID != consts.BandReturnDialogCallbackID {
		return nil, fmt.Errorf("unexpected return dialog callback id %s", request.CallbackID)
	}

	var submission models.BandReturnSubmission
	if err := jsoniter.Unmarshal(request.Submission, &submission); err != nil {
		return nil, customerrors.NewInteractiveDialogValidationError(map[string]string{consts.BandCommentField: consts.ValidationErrorInvalidFormat})
	}

	if submission.ReturnToStatusID == "" {
		return nil, customerrors.NewInteractiveDialogValidationError(map[string]string{consts.BandReturnToStatusIDField: consts.ValidationErrorFieldRequired})
	}

	comment := strings.TrimSpace(submission.Comment)
	if comment == "" {
		return nil, customerrors.NewInteractiveDialogValidationError(map[string]string{consts.BandCommentField: consts.ValidationErrorFieldRequired})
	}
	if utf8.RuneCountInString(comment) > consts.MaxReturnCommentLength {
		return nil, customerrors.NewInteractiveDialogValidationError(map[string]string{consts.BandCommentField: fmt.Sprintf(consts.ValidationErrorMaxLengthFormat, consts.MaxReturnCommentLength)})
	}

	return &models.BandActionRequest{
		TicketID:         state.TicketID,
		CategoryID:       state.CategoryID,
		StatusID:         state.StatusID,
		Action:           state.Action,
		Operation:        state.Operation,
		TypeOfEmployeeID: state.TypeOfEmployeeID,
		UserID:           request.UserID,
		EmployeeID:       state.EmployeeID,
		ChannelID:        state.ChannelID,
		PostID:           state.PostID,
		Comment:          comment,
		ReturnToStatusID: submission.ReturnToStatusID,
	}, nil
}
