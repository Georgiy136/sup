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

func (t *TicketNotificationsService) handleReturnedOperation(ctx context.Context, notification ticketmodels.TicketNotification) error {
	msgText := fmt.Sprintf(formatMsgTextTicketStatusReturned,
		notification.TicketID,
		notification.TicketInfo.TicketName,
		notification.Comments,
	)

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

	statusText := templates.GetTicketStatusTextByOperationType(consts.ReturnedOperation, notification.TicketInfo.ResponsibleEmployeeID)
	for _, post := range posts {
		if err = t.updatePostMessage(ctx, post, statusText); err != nil {
			logrus.Errorf("can't update band post %s, employee_id: %d, for ticket %d: %v", post.PostID, notification.EmployeeID, notification.TicketID, err)
		}
	}

	if _, err = t.sendNotification(ctx, notification, msgText, true); err != nil {
		return fmt.Errorf("can't send returned notification: %w", err)
	}
	return nil
}
