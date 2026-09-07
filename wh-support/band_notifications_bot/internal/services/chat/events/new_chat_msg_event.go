package events

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat"
	chatmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/employee_tags"
)

const maxNotificationAge = 5 * time.Minute

type newChatMessageEventNotifier struct {
	presence      presenceChecker
	lastChatViews lastChatViewProvider
	messenger     messenger
	modeChecker   modeChecker
}

func NewChatMessageEventNotifier(
	presence presenceChecker,
	lastChatViews lastChatViewProvider,
	messenger messenger,
	modeChecker modeChecker,
) *newChatMessageEventNotifier {
	return &newChatMessageEventNotifier{
		presence:      presence,
		lastChatViews: lastChatViews,
		messenger:     messenger,
		modeChecker:   modeChecker,
	}
}

func (h *newChatMessageEventNotifier) Event() string {
	return consts.EventNotificationNewChatMessage
}

func (h *newChatMessageEventNotifier) Handle(ctx context.Context, publications []models.Publication) error {
	notifications := h.parseNotifications(publications)
	latestByTicket := groupLatestByTicket(notifications)
	lastViewsByTicketEmployee := h.getLastViews(ctx, latestByTicket)

	for _, notification := range latestByTicket {
		if h.modeChecker.IsTestMode() {
			if !h.modeChecker.IsAllowedInTestMode(notification.SenderEmployeeID, h.modeChecker.GetTestEmployees()) {
				continue
			}
		}
		if err := h.processTicketNotification(ctx, notification, lastViewsByTicketEmployee); err != nil {
			logrus.Errorf("can't process ticket notification, err: %v", err)
		}
	}
	return nil
}

func (h *newChatMessageEventNotifier) parseNotifications(publications []models.Publication) []chatmodels.ChatEventNotification {
	result := make([]chatmodels.ChatEventNotification, 0, len(publications))

	for idx := range publications {
		var notification chatmodels.ChatEventNotification
		if err := jsoniter.Unmarshal(publications[idx].Data, &notification); err != nil {
			logrus.Errorf("can't parse band notification, offset: %d, err: %v", publications[idx].Offset, err)
			continue
		}

		createdDt, err := time.Parse(time.RFC3339Nano, notification.CreatedDt)
		if err != nil {
			logrus.Errorf("can't parse created_dt, offset: %d, err: %v", publications[idx].Offset, err)
			continue
		}
		if time.Since(createdDt) > maxNotificationAge {
			logrus.Infof("notification too old, event: %s, created_dt: %v", notification.Event, createdDt)
			continue
		}

		notification.TaggedEmployeeIDs = employee_tags.GetTaggedEmployeeIDs(notification.Message)

		result = append(result, notification)
	}
	return result
}

func (h *newChatMessageEventNotifier) processTicketNotification(ctx context.Context, event chatmodels.ChatEventNotification, lastViews map[chatmodels.LastChatViewRequest]time.Time) error {
	onlineUsers, err := h.getChatOnlineUsers(ctx, event.TicketID)
	if err != nil {
		return fmt.Errorf("event: %s, can't get ticket %d presence: %v", event.Event, event.TicketID, err)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, event.CreatedDt)
	if err != nil {
		return fmt.Errorf("event: %s, can't parse created_dt for ticket %d: %v", event.Event, event.TicketID, err)
	}

	taggedEmployees := make(map[int64]struct{}, len(event.TaggedEmployeeIDs))
	for _, employeeID := range event.TaggedEmployeeIDs {
		taggedEmployees[employeeID] = struct{}{}
	}

	notifyEmployees := make(map[int64]struct{}, len(event.NotifyEmployeeIDs))
	for _, employeeID := range event.NotifyEmployeeIDs {
		if employeeID == event.SenderEmployeeID {
			continue
		}
		if _, ok := taggedEmployees[employeeID]; ok {
			continue
		}
		notifyEmployees[employeeID] = struct{}{}
	}

	employeeLabels := make(map[int64]string, len(taggedEmployees))
	for employeeID := range taggedEmployees {
		username, err := h.messenger.GetUsernameByEmployeeID(ctx, employeeID)
		if err != nil {
			logrus.Errorf("event: %s, GetUsernameByEmployeeID failed, ticket %d, employee_id: %d, err: %v", event.Event, event.TicketID, employeeID, err)
			continue
		}
		employeeLabels[employeeID] = chat.FormatBandUsername(username)
	}

	msg := chat.BuildNotificationMessage(event.TicketID)

	attachment := chat.BuildAttachment(chatmodels.BandAttachmentParams{
		TicketID:             event.TicketID,
		Message:              event.Message,
		SenderEmployeeName:   event.SenderEmployeeName,
		CreatedDt:            event.CreatedDt,
		EmployeeLabels:       employeeLabels,
		DefaultEmployeeLabel: chat.FormatBandUsername(chat.UnknownEmployeeName),
	})

	for employeeID := range notifyEmployees {
		if err = h.sendChatNotification(
			ctx,
			event,
			employeeID,
			msg,
			attachment,
			onlineUsers,
			lastViews,
			createdAt,
		); err != nil {
			logrus.Errorf("can't send notification for ticket %d: %v", event.TicketID, err)
		}
	}

	msg = chat.BuildTagNotificationMessage(event.TicketID)

	for employeeID := range taggedEmployees {
		if employeeID == event.SenderEmployeeID {
			continue
		}
		if err = h.sendChatNotification(
			ctx,
			event,
			employeeID,
			msg,
			attachment,
			onlineUsers,
			lastViews,
			createdAt,
		); err != nil {
			logrus.Errorf("can't send notification for ticket %d: %v", event.TicketID, err)
		}
	}
	return nil
}

func (h *newChatMessageEventNotifier) sendChatNotification(ctx context.Context, event chatmodels.ChatEventNotification, employeeID int64, msg string, attachment models.BandAttachment, onlineUsers map[int64]struct{}, lastViews map[chatmodels.LastChatViewRequest]time.Time, createdAt time.Time) error {
	if _, online := onlineUsers[employeeID]; online {
		logrus.Infof("skip event: %s, employee %d is viewing chat %d", event.Event, employeeID, event.TicketID)
		return nil
	}
	if wasMessageRead(lastViews, employeeID, event.TicketID, createdAt) {
		logrus.Infof("skip event: %s, employee %d has already read chat %d", event.Event, employeeID, event.TicketID)
		return nil
	}

	if _, err := h.messenger.SendNotification(ctx, models.BandRecipient{EmployeeID: employeeID}, msg, attachment); err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) || errors.Is(err, customerrors.ErrTooManyUsers) {
			logrus.Infof("can't send band notification to employee %d: %v", employeeID, err)
			return nil
		}
		return fmt.Errorf("can't send band notification to employee %d: %w", employeeID, err)
	}
	logrus.Infof("band notification successfully sent for ticket №%d to employee %d", event.TicketID, employeeID)
	return nil
}

func (h *newChatMessageEventNotifier) getChatOnlineUsers(ctx context.Context, ticketID int64) (map[int64]struct{}, error) {
	result, err := h.presence.Presence(ctx, models.PresenceRequest{
		Channel: fmt.Sprintf(consts.TicketChannelFormat, ticketID),
	})
	if err != nil {
		return nil, fmt.Errorf("get presence for ticket %d: %w", ticketID, err)
	}

	onlineUsers := make(map[int64]struct{}, len(result.Presence))
	for _, client := range result.Presence {
		employeeID, err := strconv.ParseInt(client.User, 10, 64)
		if err != nil {
			logrus.Warnf("can't parse presence user id %q: %v", client.User, err)
			continue
		}
		onlineUsers[employeeID] = struct{}{}
	}
	return onlineUsers, nil
}

func (h *newChatMessageEventNotifier) getLastViews(ctx context.Context, latestByTicket map[int64]chatmodels.ChatEventNotification) map[chatmodels.LastChatViewRequest]time.Time {
	lastChatViewReq := make([]chatmodels.LastChatViewRequest, 0, len(latestByTicket))
	for _, notification := range latestByTicket {
		for _, employeeID := range notification.NotifyEmployeeIDs {
			lastChatViewReq = append(lastChatViewReq, chatmodels.LastChatViewRequest{EmployeeID: employeeID, TicketID: notification.TicketID})
		}
		for _, employeeID := range notification.TaggedEmployeeIDs {
			lastChatViewReq = append(lastChatViewReq, chatmodels.LastChatViewRequest{EmployeeID: employeeID, TicketID: notification.TicketID})
		}
	}

	lastViews, err := h.lastChatViews.GetLastChatViewsByUsers(ctx, lastChatViewReq)
	if err != nil {
		logrus.Errorf("can't get last chat views: %v", err)
		return map[chatmodels.LastChatViewRequest]time.Time{}
	}
	return lastViews
}

func groupLatestByTicket(notifications []chatmodels.ChatEventNotification) map[int64]chatmodels.ChatEventNotification {
	latestByTicket := make(map[int64]chatmodels.ChatEventNotification, len(notifications))
	for i := range notifications {
		latestByTicket[notifications[i].TicketID] = notifications[i]
	}
	return latestByTicket
}

func wasMessageRead(lastViews map[chatmodels.LastChatViewRequest]time.Time, employeeID, ticketID int64, createdAt time.Time) bool {
	if viewedAt, ok := lastViews[chatmodels.LastChatViewRequest{EmployeeID: employeeID, TicketID: ticketID}]; ok {
		if viewedAt.After(createdAt) {
			return true
		}
	}
	return false
}

type presenceChecker interface {
	Presence(ctx context.Context, req models.PresenceRequest) (*models.PresenceResult, error)
}

type lastChatViewProvider interface {
	GetLastChatViewsByUsers(ctx context.Context, requests []chatmodels.LastChatViewRequest) (map[chatmodels.LastChatViewRequest]time.Time, error)
}

type messenger interface {
	SendNotification(ctx context.Context, recipient models.BandRecipient, msgNotification string, attachments ...models.BandAttachment) (*models.BandSentNotification, error)
	GetUsernameByEmployeeID(ctx context.Context, employeeID int64) (string, error)
}

type modeChecker interface {
	IsTestMode() bool
	IsAllowedInTestMode(employeeID int64, employeesList []int64) bool
	GetTestEmployees() []int64
}
