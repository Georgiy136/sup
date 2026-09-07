package ticket

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/templates"
)

func (t *TicketNotificationsService) handleApproveOperation(ctx context.Context, notification ticketmodels.TicketNotification, requiredSiteAction bool) error {
	msgText, err := t.notificationBuilder.GenMsgNotification(notification, requiredSiteAction)
	if err != nil {
		return fmt.Errorf("can't gen notifications message: %w", err)
	}

	sentPost, err := t.sendNotification(ctx, notification, msgText, requiredSiteAction)
	if err != nil {
		return fmt.Errorf("can't send notifications message: %w", err)
	}

	if sentPost == nil || sentPost.PostID == "" {
		logrus.Infof("can't save band post for ticket %d, employee %d: empty post id", notification.TicketID, notification.EmployeeID)
		return nil
	}
	return t.ticketPostCache.SaveBandTicketPost(ctx, ticketmodels.TicketPost{
		TicketID:         notification.TicketID,
		CategoryID:       notification.TicketInfo.CategoryID,
		PostID:           sentPost.PostID,
		ChannelID:        sentPost.ChannelID,
		Operation:        notification.TicketInfo.OperationType,
		TypeOfEmployeeID: notification.TypeOfEmployeeID,
		EmployeeID:       notification.EmployeeID,
	})
}

func (t *TicketNotificationsService) handleApprovedOperation(ctx context.Context, notification ticketmodels.TicketNotification) error {
	if notification.TypeOfEmployeeID.IsCreator() {
		if err := t.sendTicketMsgForCreator(ctx, notification); err != nil {
			return fmt.Errorf("send ticket for creator error: %w", err)
		}
	}

	channelID, err := t.bandBot.GetDirectChannelID(ctx, notification.EmployeeID)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) || errors.Is(err, customerrors.ErrTooManyUsers) {
			logrus.Infof("can't get direct channel_id, ticket %d; %v", notification.TicketID, err)
			return nil
		}
		return fmt.Errorf("can't get channel id for ticket %d: %w", notification.TicketID, err)
	}

	posts, err := t.ticketPostCache.GetBandTicketPostsByChannel(ctx, notification.TicketID, channelID)
	if err != nil {
		return fmt.Errorf("can't get saved band posts: %w", err)
	}

	statusText := templates.GetTicketStatusTextByOperationType(consts.ApprovedOperation, notification.TicketInfo.ResponsibleEmployeeID)
	for _, post := range posts {
		if post.Operation != consts.ApproveOperation {
			continue
		}
		if err = t.updatePostMessage(ctx, post, statusText); err != nil {
			logrus.Errorf("can't update band post %s, employee_id: %d, for ticket %d: %v", post.PostID, notification.EmployeeID, notification.TicketID, err)
		}
	}

	return t.ticketPostCache.DeleteBandTicketPostsByOperation(ctx, notification.TicketID, channelID, consts.ApproveOperation)
}
