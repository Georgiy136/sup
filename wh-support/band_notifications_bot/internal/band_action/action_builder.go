package bandaction

import (
	"fmt"
	"strconv"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
)

func (b *BandActionService) BuildApproveActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction {
	actions := make([]models.BandAction, 0, 3)

	if !opts.BlockStatus {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.ApproveAction, consts.ApproveOperation, consts.BandApproveButtonName, consts.BandStyleSuccess))
	}
	if !opts.BlockReject {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.RejectAction, consts.ApproveOperation, consts.BandRejectButtonName, consts.BandStyleDanger))
	}
	if opts.HasReturnStatuses {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.ReturnAction, consts.ApproveOperation, consts.BandReturnButtonName, consts.BandStyleWarning))
	}

	return actions
}

func (b *BandActionService) BuildBookActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction {
	actions := make([]models.BandAction, 0, 3)

	if !opts.BlockStatus {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.BookAction, consts.PerformOperation, consts.BandBookButtonName, consts.BandStylePrimary))
	}
	if !opts.BlockReject {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.RejectAction, consts.PerformOperation, consts.BandRejectButtonName, consts.BandStyleDanger))
	}
	if opts.HasReturnStatuses {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.ReturnAction, consts.PerformOperation, consts.BandReturnButtonName, consts.BandStyleWarning))
	}

	return actions
}

func (b *BandActionService) BuildPerformActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction {
	actionCtx.TypeOfEmployeeID = ticketmodels.TypeOfEmployeeEmployee
	actions := make([]models.BandAction, 0, 4)

	if !opts.BlockStatus {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.PerformAction, consts.PerformOperation, consts.BandPerformButtonName, consts.BandStyleSuccess))
	}
	actions = append(actions, b.buildTicketAction(actionCtx, consts.UnbookAction, consts.PerformOperation, consts.BandUnbookButtonName, consts.BandStyleWarning))
	if !opts.BlockReject {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.RejectAction, consts.PerformOperation, consts.BandRejectButtonName, consts.BandStyleDanger))
	}
	if opts.HasReturnStatuses {
		actions = append(actions, b.buildTicketAction(actionCtx, consts.ReturnAction, consts.PerformOperation, consts.BandReturnButtonName, consts.BandStyleWarning))
	}

	return actions
}

func (b *BandActionService) buildTicketAction(actionCtx models.BandActionContext, action, operation, name, style string) models.BandAction {
	now := time.Now().Unix()
	return models.BandAction{
		ID:    fmt.Sprintf("ticket%d%s", actionCtx.TicketID, action),
		Name:  name,
		Style: style,
		Path:  consts.BandActionPath,
		Context: map[string]any{
			"ticket_id":           actionCtx.TicketID,
			"action":              action,
			"operation":           operation,
			"employee_id":         actionCtx.EmployeeID,
			"type_of_employee_id": actionCtx.TypeOfEmployeeID.String(),
			"category_id":         actionCtx.CategoryID,
			"status_id":           actionCtx.StatusID,
			"ts":                  now,
			"sig": b.actionSigner.Sign(
				strconv.FormatInt(actionCtx.TicketID, 10),
				action,
				operation,
				strconv.FormatInt(actionCtx.EmployeeID, 10),
				actionCtx.TypeOfEmployeeID.String(),
				strconv.FormatInt(now, 10),
			),
		},
	}
}
