package bandaction

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/templates"
)

func (b *BandActionService) HandleBandAction(ctx context.Context, req models.BandActionRequest) error {
	if err := b.validateBandActionRequest(ctx, req); err != nil {
		return fmt.Errorf("can't validate band action request: %w", err)
	}

	switch req.Action {
	case consts.ReturnAction:
		if err := b.openReturnToStatusDialog(ctx, req); err != nil {
			return fmt.Errorf("can't open return to status dialog: %w", err)
		}
		return nil
	case consts.RejectAction:
		if err := b.openRejectDialog(ctx, req); err != nil {
			return fmt.Errorf("can't open reject dialog: %w", err)
		}
		return nil
	case consts.ApproveAction, consts.PerformAction:
		categoryStatusResp, err := b.ticketRepo.GetCategoryStatusByStatus(ctx, req.CategoryID, req.StatusID)
		if err != nil {
			return fmt.Errorf("can't get category status  from db for ticket %d, category_id: %d, status_id: %s, error: %w", req.TicketID, req.CategoryID, req.StatusID, err)
		}
		if len(categoryStatusResp.Data) != 1 {
			return fmt.Errorf("invalid length data from resp db: %d, expected 1", len(categoryStatusResp.Data))
		}
		if hasFieldsToFill(&categoryStatusResp.Data[0]) {
			if err := b.openAdditionalInfoDialogForAction(ctx, req, &categoryStatusResp.Data[0]); err != nil {
				if err := b.bandBot.SendChannelMessage(ctx, req.ChannelID, templates.ConfirmOnSite); err != nil {
					return fmt.Errorf("can't send channel message: %w", err)
				}
				return fmt.Errorf("can't handle action, open additional info dialog for action, error: %w", err)
			}
			return nil
		}
		if err = b.executeAction(ctx, req, nil); err != nil {
			return fmt.Errorf("can't execute action %s: %w", req.Action, err)
		}
		return nil
	case consts.BookAction, consts.UnbookAction:
		if err := b.executeAction(ctx, req, nil); err != nil {
			return fmt.Errorf("can't execute action %s: %w", req.Action, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown action %s: %w", req.Action, customerrors.ErrUnknownAction)
	}
}

func (b *BandActionService) HandleBandRejectDialog(ctx context.Context, req models.BandActionRequest) error {
	if err := b.validateBandActionRequest(ctx, req); err != nil {
		return fmt.Errorf("can't validate band action request: %w", err)
	}
	if err := b.executeAction(ctx, req, nil); err != nil {
		return fmt.Errorf("can't execute action %s: %w", req.Action, err)
	}
	return nil
}

func (b *BandActionService) HandleBandReturnToStatusDialog(ctx context.Context, req models.BandActionRequest) error {
	if err := b.validateBandActionRequest(ctx, req); err != nil {
		return fmt.Errorf("can't validate band action request: %w", err)
	}
	if err := b.executeAction(ctx, req, nil); err != nil {
		return fmt.Errorf("can't execute action %s: %w", req.Action, err)
	}
	return nil
}

func (b *BandActionService) HandleBandAdditionalInfoDialog(ctx context.Context, req models.BandActionRequest, filledFields map[string]any, state models.BandDialogState) error {
	if err := b.validateBandActionRequest(ctx, req); err != nil {
		return fmt.Errorf("can't validate band action request: %w", err)
	}
	if req.Action != consts.ApproveAction && req.Action != consts.PerformAction {
		return fmt.Errorf("unknown action %s for fields dialog: %w", req.Action, customerrors.ErrUnknownAction)
	}

	values, err := b.validateAndConvertSubmissionValues(ctx, state.CategoryID, state.StatusID, filledFields)
	if err != nil {
		return fmt.Errorf("can't validate filled fields %s: %w", req.UserID, err)
	}
	if err = b.executeAction(ctx, req, values); err != nil {
		return fmt.Errorf("can't execute action %s: %w", req.Action, err)
	}
	return nil
}

func (b *BandActionService) validateBandActionRequest(ctx context.Context, req models.BandActionRequest) error {
	if err := validateBandActionFields(req); err != nil {
		return fmt.Errorf("can't validate band action fields: %w", err)
	}
	employeeID, err := b.bandBot.GetEmployeeIDByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("can't get employee id by band user %s: %w", req.UserID, err)
	}
	if req.EmployeeID != employeeID {
		return fmt.Errorf("employee_id mismatch for user_id %s, want %d, got %d", req.UserID, req.EmployeeID, employeeID)
	}
	return nil
}

func validateBandActionFields(req models.BandActionRequest) error {
	if req.TicketID <= 0 {
		return fmt.Errorf("invalid ticket id %d: %w", req.TicketID, customerrors.ErrInvalidAction)
	}
	if req.UserID == "" {
		return fmt.Errorf("empty band user id: %w", customerrors.ErrInvalidAction)
	}
	if req.PostID == "" {
		return fmt.Errorf("empty band post id: %w", customerrors.ErrInvalidAction)
	}
	if req.Action == "" {
		return fmt.Errorf("empty band action: %w", customerrors.ErrInvalidAction)
	}
	if req.Operation == "" {
		return fmt.Errorf("empty band operation: %w", customerrors.ErrInvalidAction)
	}
	if req.TypeOfEmployeeID.IsEmpty() {
		return fmt.Errorf("empty band type_of_employee_id: %w", customerrors.ErrInvalidAction)
	}
	if !isActionAllowedForOperation(req.Operation, req.TypeOfEmployeeID, req.Action) {
		return fmt.Errorf("action %s is not allowed for operation %s and type_of_employee_id %s: %w", req.Action, req.Operation, req.TypeOfEmployeeID, customerrors.ErrInvalidAction)
	}
	return nil
}

func isActionAllowedForOperation(operation string, typeOfEmployeeID ticketmodels.TypeOfEmployee, action string) bool {
	switch operation {
	case consts.ApproveOperation:
		return (action == consts.ApproveAction || action == consts.RejectAction || action == consts.ReturnAction) && (typeOfEmployeeID.IsGroup() || typeOfEmployeeID.IsEmployee())
	case consts.PerformOperation:
		switch action {
		case consts.BookAction, consts.RejectAction, consts.ReturnAction:
			return typeOfEmployeeID.IsGroup() || typeOfEmployeeID.IsEmployee()
		case consts.PerformAction, consts.UnbookAction:
			return typeOfEmployeeID.IsEmployee()
		}
	}
	return false
}
