package ticket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/templates"
)

func (t *TicketNotificationsService) handleBookedOperation(ctx context.Context, notification ticketmodels.TicketNotification) error {
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

	categoryStatusData, err := t.ticketRepo.GetCategoryStatusByStatus(ctx, notification.TicketInfo.CategoryID, notification.TicketInfo.StatusID)
	if err != nil {
		return fmt.Errorf("can't check category status for ticket %d: %v", notification.TicketID, err)
	}
	if categoryStatusData == nil || len(categoryStatusData.Data) == 0 {
		return fmt.Errorf("can't handle booked operation, category status data len is zero, ticket_id %d, category_id %d, status_id %s,", notification.TicketID, notification.TicketInfo.CategoryID, notification.TicketInfo.StatusID)
	}

	statusText := templates.GetTicketStatusTextByOperationType(consts.BookedOperation, notification.TicketInfo.ResponsibleEmployeeID)
	for _, post := range posts {
		messageText, err := t.bandBot.GetPostMessage(ctx, post.PostID)
		if err != nil {
			logrus.Errorf("can't get band post %s message for ticket %d: %v", post.PostID, notification.TicketID, err)
			continue
		}

		if notification.TypeOfEmployeeID.IsEmployee() {
			updateMessage := messageText
			var attachments []models.BandAttachment
			if t.bandActions.CanAttachBandActions(&categoryStatusData.Data[0]) {
				attachments = []models.BandAttachment{{Actions: t.bandActions.BuildPerformActions(models.BandActionContext{
					TicketID:         notification.TicketID,
					EmployeeID:       notification.EmployeeID,
					CategoryID:       notification.TicketInfo.CategoryID,
					StatusID:         notification.TicketInfo.StatusID,
					TypeOfEmployeeID: post.TypeOfEmployeeID,
				}, ticketmodels.ActionOptions{
					BlockReject:       notification.TicketInfo.IsBlockedReject,
					BlockStatus:       notification.TicketInfo.IsBlockedStatus,
					HasReturnStatuses: len(notification.TicketInfo.ReturnStatuses) > 0,
				})}}

			} else if !strings.HasSuffix(messageText, templates.PerformOnSiteTicketMessage) {
				updateMessage += templates.PerformOnSiteTicketMessage
			}

			if err = t.bandBot.UpdatePost(ctx, post.PostID, updateMessage, attachments...); err != nil {
				logrus.Errorf("can't update band post %s, employee_id: %d, for ticket %d: %v", post.PostID, notification.EmployeeID, notification.TicketID, err)
			}
			continue
		}

		if !strings.HasSuffix(messageText, statusText) {
			messageText += statusText
		}
		if err = t.bandBot.UpdatePost(ctx, post.PostID, messageText); err != nil {
			logrus.Errorf("can't update band post %s, employee_id: %d, for ticket %d: %v", post.PostID, notification.EmployeeID, notification.TicketID, err)
		}
	}
	return nil
}

func (t *TicketNotificationsService) handleUnbookedOperation(ctx context.Context, notification ticketmodels.TicketNotification) error {
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

	bookedStatusText := templates.GetTicketStatusTextByOperationType(consts.BookedOperation, notification.TicketInfo.ResponsibleEmployeeID)
	for _, post := range posts {
		var attachments []models.BandAttachment
		if post.TypeOfEmployeeID.IsGroup() || post.TypeOfEmployeeID.IsEmployee() {
			attachments = []models.BandAttachment{{Actions: t.bandActions.BuildBookActions(models.BandActionContext{
				TicketID:         notification.TicketID,
				EmployeeID:       notification.EmployeeID,
				CategoryID:       notification.TicketInfo.CategoryID,
				StatusID:         notification.TicketInfo.StatusID,
				TypeOfEmployeeID: post.TypeOfEmployeeID,
			}, ticketmodels.ActionOptions{
				BlockReject:       notification.TicketInfo.IsBlockedReject,
				BlockStatus:       notification.TicketInfo.IsBlockedStatus,
				HasReturnStatuses: len(notification.TicketInfo.ReturnStatuses) > 0,
			})}}

		}

		messageText, err := t.bandBot.GetPostMessage(ctx, post.PostID)
		if err != nil {
			logrus.Errorf("can't get band post %s message for ticket %d: %v", post.PostID, notification.TicketID, err)
			continue
		}

		if before, _, found := strings.Cut(messageText, bookedStatusText); found {
			messageText = before
		}
		if err = t.bandBot.UpdatePost(ctx, post.PostID, messageText, attachments...); err != nil {
			logrus.Errorf("can't update band post %s, employee_id: %d, for ticket %d: %v", post.PostID, notification.EmployeeID, notification.TicketID, err)

		}
	}
	return nil
}
