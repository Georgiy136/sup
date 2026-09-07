package bandaction

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/templates"
)

func (b *BandActionService) executeAction(ctx context.Context, req models.BandActionRequest, values map[string]any) error {
	statusText, err := b.processAction(ctx, req, values)
	if err != nil {
		if sendErr := b.sendUserActionError(ctx, req, err); sendErr != nil {
			return fmt.Errorf("send user action error: %w", sendErr)
		}
		if isDBActionStatusError(err, http.StatusUnprocessableEntity) {
			logrus.Warnf("process action business error, req: %+v, error: %v", req, err)
			return nil
		}
		logrus.Infof("process action error, req: %+v, error: %v", req, err)
		return fmt.Errorf("process action error: %w", err)
	}

	messageText, err := b.bandBot.GetPostMessage(ctx, req.PostID)
	if err != nil {
		return fmt.Errorf("can't get band post %s message: %w", req.PostID, err)
	}
	messageText += statusText

	if err = b.updateBandPostAfterAction(ctx, req, messageText, req.EmployeeID); err != nil {
		return fmt.Errorf("can't update band post %s: %w", req.PostID, err)
	}
	if err = b.cleanupSavedPosts(ctx, req); err != nil {
		return fmt.Errorf("can't cleanup saved posts for ticket %d: %w", req.TicketID, err)
	}
	return nil
}

func (b *BandActionService) processAction(ctx context.Context, req models.BandActionRequest, values map[string]any) (string, error) {
	switch req.Action {
	case consts.ApproveAction:
		if err := b.ticketRepo.Approve(ctx, req.TicketID, req.EmployeeID, values, 0); err != nil {
			return "", fmt.Errorf("can't approve ticket %d: %w", req.TicketID, err)
		}
		return templates.GetTicketStatusTextByOperationType(consts.ApprovedOperation, &req.EmployeeID), nil
	case consts.BookAction:
		if err := b.ticketRepo.Book(ctx, req.TicketID, req.EmployeeID); err != nil {
			return "", fmt.Errorf("can't book ticket %d: %w", req.TicketID, err)
		}
		return templates.GetTicketStatusTextByOperationType(consts.BookedOperation, &req.EmployeeID), nil
	case consts.UnbookAction:
		if err := b.ticketRepo.Unbook(ctx, req.TicketID, req.EmployeeID); err != nil {
			return "", fmt.Errorf("can't unbook ticket %d: %w", req.TicketID, err)
		}
		return "", nil
	case consts.PerformAction:
		if err := b.ticketRepo.Perform(ctx, req.TicketID, req.EmployeeID, values, 0); err != nil {
			return "", fmt.Errorf("can't perform ticket %d: %w", req.TicketID, err)
		}
		return templates.GetTicketStatusTextByOperationType(consts.PerformedOperation, &req.EmployeeID), nil
	case consts.RejectAction:
		if req.Comment == "" {
			return "", fmt.Errorf("empty reject comment: %w", customerrors.ErrInvalidAction)
		}
		if err := b.ticketRepo.Reject(ctx, req.TicketID, req.EmployeeID, req.Comment); err != nil {
			return "", fmt.Errorf("can't reject ticket %d: %w", req.TicketID, err)
		}
		return templates.GetTicketStatusTextByOperationType(consts.RejectOperation, &req.EmployeeID), nil
	case consts.ReturnAction:
		if req.Comment == "" {
			return "", fmt.Errorf("empty return comment: %w", customerrors.ErrInvalidAction)
		}
		if req.ReturnToStatusID == "" {
			return "", fmt.Errorf("empty return status: %w", customerrors.ErrInvalidAction)
		}
		if err := b.ticketRepo.ReturnToStatus(ctx, req.TicketID, req.ReturnToStatusID, req.Comment, req.EmployeeID); err != nil {
			return "", fmt.Errorf("can't return ticket %d: %w", req.TicketID, err)
		}
		return templates.GetTicketStatusTextByOperationType(consts.ReturnedOperation, &req.EmployeeID), nil
	default:
		return "", fmt.Errorf("unknown action %s: %w", req.Action, customerrors.ErrUnknownAction)
	}
}

func (b *BandActionService) updateBandPostAfterAction(ctx context.Context, req models.BandActionRequest, messageText string, employeeID int64) error {
	switch req.Action {
	case consts.BookAction:
		categoryStatusData, err := b.ticketRepo.GetCategoryStatusByStatus(ctx, req.CategoryID, req.StatusID)
		if err != nil {
			return fmt.Errorf("can't check category status for ticket %d: %v", req.TicketID, err)
		}
		if categoryStatusData == nil || len(categoryStatusData.Data) == 0 {
			return fmt.Errorf("can't handle book action, category status data len is zero, ticket_id %d, category_id %d, status_id %s", req.TicketID, req.CategoryID, req.StatusID)
		}
		updateMessage := messageText

		var attachments []models.BandAttachment
		if b.CanAttachBandActions(&categoryStatusData.Data[0]) {
			attachments = []models.BandAttachment{{Actions: b.BuildPerformActions(models.BandActionContext{
				TicketID:   req.TicketID,
				EmployeeID: req.EmployeeID,
				CategoryID: req.CategoryID,
				StatusID:   req.StatusID,
			}, ticketmodels.ActionOptions{
				BlockReject:       categoryStatusData.Data[0].IsBlockedReject,
				BlockStatus:       categoryStatusData.Data[0].IsBlockedStatus,
				HasReturnStatuses: len(categoryStatusData.Data[0].ReturnStatuses) > 0,
			})}}
		} else {
			updateMessage += templates.PerformOnSiteTicketMessage
		}

		return b.bandBot.UpdatePost(ctx, req.PostID, updateMessage, attachments...)
	case consts.UnbookAction:
		bookedStatusText := templates.GetTicketStatusTextByOperationType(consts.BookedOperation, &employeeID)
		if before, _, found := strings.Cut(messageText, bookedStatusText); found {
			messageText = before
		}
		categoryStatusData, err := b.ticketRepo.GetCategoryStatusByStatus(ctx, req.CategoryID, req.StatusID)
		if err != nil {
			return fmt.Errorf("can't check category status for ticket %d: %v", req.TicketID, err)
		}
		if categoryStatusData == nil || len(categoryStatusData.Data) == 0 {
			return fmt.Errorf("can't handle unbook action, category status data len is zero, ticket_id %d, category_id %d, status_id %s", req.TicketID, req.CategoryID, req.StatusID)
		}
		return b.bandBot.UpdatePost(ctx, req.PostID, messageText,
			models.BandAttachment{Actions: b.BuildBookActions(models.BandActionContext{
				TicketID:         req.TicketID,
				EmployeeID:       req.EmployeeID,
				CategoryID:       req.CategoryID,
				StatusID:         req.StatusID,
				TypeOfEmployeeID: req.TypeOfEmployeeID,
			}, ticketmodels.ActionOptions{
				BlockReject:       categoryStatusData.Data[0].IsBlockedReject,
				BlockStatus:       categoryStatusData.Data[0].IsBlockedStatus,
				HasReturnStatuses: len(categoryStatusData.Data[0].ReturnStatuses) > 0,
			})})
	case consts.ApproveAction, consts.PerformAction, consts.RejectAction, consts.ReturnAction:
		return b.bandBot.UpdatePost(ctx, req.PostID, messageText)
	}
	return nil
}

func (b *BandActionService) cleanupSavedPosts(ctx context.Context, req models.BandActionRequest) error {
	switch req.Action {
	case consts.ApproveAction, consts.PerformAction:
		return b.ticketPostCache.DeleteBandTicketPostsByOperation(ctx, req.TicketID, req.ChannelID, req.Operation)
	case consts.RejectAction, consts.ReturnAction:
		return b.ticketPostCache.DeleteBandTicketPosts(ctx, req.TicketID, req.ChannelID)
	}
	return nil
}

func (b *BandActionService) sendUserActionError(ctx context.Context, req models.BandActionRequest, actionErr error) error {
	message := templates.ActionInternalErrorMessage
	if dbErr, ok := errors.AsType[*customerrors.DBActionError](actionErr); ok && dbErr.StatusCode == http.StatusUnprocessableEntity {
		message = dbErr.Message
	}
	return b.bandBot.SendChannelMessage(ctx, req.ChannelID, fmt.Sprintf(templates.ActionErrorMessageFormat, req.TicketID, message))
}

func isDBActionStatusError(err error, statusCode int) bool {
	if dbErr, ok := errors.AsType[*customerrors.DBActionError](err); ok {
		return dbErr.StatusCode == statusCode
	}
	return false
}
