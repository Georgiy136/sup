package ticket

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"

	"github.com/sirupsen/logrus"
)

func (t *TicketNotificationsService) sendNotification(ctx context.Context, notification ticketmodels.TicketNotification, msgNotification string, requiredSiteAction bool) (*models.BandSentNotification, error) {
	var actions []models.BandAction

	if !requiredSiteAction {
		actions = t.actionsForNotification(notification)
	}

	sentNotification, err := t.bandBot.SendNotification(ctx, models.BandRecipient{EmployeeID: notification.EmployeeID}, msgNotification, models.BandAttachment{Actions: actions})
	if err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) || errors.Is(err, customerrors.ErrTooManyUsers) {
			logrus.Infof("can't send notification on ticket %d, employee %d: %v", notification.TicketID, notification.EmployeeID, err)
			return nil, nil
		}
		return nil, fmt.Errorf("can't send notification: %w", err)
	}
	return sentNotification, nil
}

func (t *TicketNotificationsService) actionsForNotification(notification ticketmodels.TicketNotification) []models.BandAction {
	if !notification.TypeOfEmployeeID.IsGroup() && !notification.TypeOfEmployeeID.IsEmployee() {
		return nil
	}

	actionCtx := models.BandActionContext{
		TicketID:         notification.TicketID,
		EmployeeID:       notification.EmployeeID,
		CategoryID:       notification.TicketInfo.CategoryID,
		StatusID:         notification.TicketInfo.StatusID,
		TypeOfEmployeeID: notification.TypeOfEmployeeID,
	}
	opts := ticketmodels.ActionOptions{
		BlockReject:       notification.TicketInfo.IsBlockedReject,
		BlockStatus:       notification.TicketInfo.IsBlockedStatus,
		HasReturnStatuses: len(notification.TicketInfo.ReturnStatuses) > 0,
	}

	switch notification.TicketInfo.OperationType {
	case consts.ApproveOperation:
		return t.bandActions.BuildApproveActions(actionCtx, opts)
	case consts.PerformOperation:
		return t.bandActions.BuildBookActions(actionCtx, opts)
	}
	return nil
}

func (t *TicketNotificationsService) updatePostMessage(ctx context.Context, post ticketmodels.TicketPost, statusText string) error {
	messageText, err := t.bandBot.GetPostMessage(ctx, post.PostID)
	if err != nil {
		return fmt.Errorf("can't get band post %s message: %w", post.PostID, err)
	}
	return t.bandBot.UpdatePost(ctx, post.PostID, messageText+statusText)
}

func (t *TicketNotificationsService) sendTicketMsgForCreator(ctx context.Context, notification ticketmodels.TicketNotification) error {
	msg := t.notificationBuilder.GenShortMsgTextForTicketCreator(notification)
	if _, err := t.bandBot.SendNotification(ctx, models.BandRecipient{EmployeeID: notification.EmployeeID}, msg); err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) || errors.Is(err, customerrors.ErrTooManyUsers) {
			logrus.Infof("can't send notification on ticket %d, employee %d: %v", notification.TicketID, notification.EmployeeID, err)
			return nil
		}
		return fmt.Errorf("send ticket for creator error: %w", err)
	}
	return nil
}
