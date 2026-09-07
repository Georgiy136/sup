package realtime

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	maps_utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"

	"github.com/sirupsen/logrus"
)

func (r *Realtime) startProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("realtime processor stopped: %v", ctx.Err())
			return
		case event := <-r.eventCh:
			if err := r.processTicketEvent(ctx, event); err != nil {
				logrus.Errorf("ticket event processor error: %v", err)
			}
			if err := r.processUsersEvent(ctx, event); err != nil {
				logrus.Errorf("user event processor error: %v", err)
			}
		}
	}
}

func (r *Realtime) processTicketEvent(ctx context.Context, event models.RealtimePostEvent) error {
	publishData := models.CentrifugoMsgPublishData{
		Event:        "chat.message.created",
		Message:      event.Message,
		EmployeeID:   event.EmployeeID,
		EmployeeName: event.EmployeeName,
		CreatedDt:    event.CreatedDt,
		Files:        event.Files,
	}

	if err := r.centrifugo.Publish(ctx, fmt.Sprintf("ticket:%d", event.TicketID), publishData); err != nil {
		return fmt.Errorf("publish ticket event error: %w", err)
	}
	if err := r.activityStorage.SetTicketLastActivity(ctx, event.TicketID, time.Now()); err != nil {
		return fmt.Errorf("update redis last activity error: %w", err)
	}
	return nil
}

func (r *Realtime) processUsersEvent(ctx context.Context, event models.RealtimePostEvent) error {
	ticketInfo, err := r.chatStorage.GetTicketInfo(event.TicketID)
	if err != nil {
		return fmt.Errorf("get ticket info error: %w", err)
	}
	if ticketInfo == nil {
		return nil
	}

	allowedUsers := getTicketUsers(ticketInfo)
	notifyEmployeeIDs := make([]int64, 0, len(allowedUsers))
	for _, employeeID := range allowedUsers {
		if employeeID == event.EmployeeID {
			continue
		}

		if err = r.processToUserChannel(ctx, employeeID, event); err != nil {
			logrus.Errorf("process to user channel error: %v", err)
		}
	}

	notificationUsers, err := r.getTicketNotificationUsers(ctx, ticketInfo)
	if err != nil {
		logrus.Errorf("get ticket notification users error: %v", err)
	}
	for _, employeeID := range notificationUsers {
		if employeeID == event.EmployeeID {
			continue
		}
		notifyEmployeeIDs = append(notifyEmployeeIDs, employeeID)
	}

	if len(notifyEmployeeIDs) == 0 {
		return nil
	}
	if err = r.processToChatNotificationChannel(ctx, notifyEmployeeIDs, event); err != nil {
		logrus.Errorf("process to chat notification channel error: %v", err)
	}
	return nil
}

//nolint:gochecknoglobals
var ticketNotificationActionTypes = []string{
	models.WorkCategoriesApproveActionType,
	models.WorkCategoriesPerformActionType,
}

func getTicketUsers(ticketInfo *models.TicketInfo) []int64 {
	uniqueEmployeeIDs := make(map[int64]struct{})
	addEmployeeIDs(uniqueEmployeeIDs, ticketInfo.CreateEmployeeID)
	addEmployeeIDs(uniqueEmployeeIDs, ticketInfo.FavouriteEmployeeIDs...)
	addEmployeeIDs(uniqueEmployeeIDs, ticketInfo.GroupEmployeeIDs...)
	return maps_utils.Keys(uniqueEmployeeIDs)
}

func (r *Realtime) getTicketNotificationUsers(ctx context.Context, ticketInfo *models.TicketInfo) ([]int64, error) {
	uniqueEmployeeIDs := make(map[int64]struct{})
	addEmployeeIDs(uniqueEmployeeIDs, ticketInfo.CreateEmployeeID)
	addEmployeeIDs(uniqueEmployeeIDs, ticketInfo.FavouriteEmployeeIDs...)

	resourcePolicies, err := r.accessStorage.GetResourceAccessPolicies(ctx, ticketInfo.CategoryID, ticketInfo.StatusID, ticketNotificationActionTypes...)
	if err != nil {
		return nil, fmt.Errorf("get resource access policies error: %w", err)
	}

	var countEmployeesByGroupCategory int64
	for _, employeeID := range ticketInfo.GroupEmployeeIDs {
		hasAccess, err := r.hasTicketStatusAccess(ctx, employeeID, resourcePolicies)
		if err != nil {
			return nil, fmt.Errorf("check employee %d ticket status access error: %w", employeeID, err)
		}
		if hasAccess {
			prevEmployeesCount := len(uniqueEmployeeIDs)

			addEmployeeIDs(uniqueEmployeeIDs, employeeID)

			if len(uniqueEmployeeIDs) != prevEmployeesCount {
				countEmployeesByGroupCategory++
			}
		}
	}

	if countEmployeesByGroupCategory == 0 {
		logrus.Errorf("count employees in group: %d, category: %d, status: %s, ticket_id: %d", countEmployeesByGroupCategory, ticketInfo.CategoryID, ticketInfo.StatusID, ticketInfo.TicketID)
	} else {
		logrus.Infof("count employees in group: %d, category: %d, status: %s, ticket_id: %d", countEmployeesByGroupCategory, ticketInfo.CategoryID, ticketInfo.StatusID, ticketInfo.TicketID)
	}

	return maps_utils.Keys(uniqueEmployeeIDs), nil
}

func (r *Realtime) hasTicketStatusAccess(ctx context.Context, employeeID int64, resourcePolicies map[string][]int64) (bool, error) {
	employeeGroups, err := r.accessStorage.GetEmployeeAccessGroups(ctx, employeeID)
	if err != nil {
		return false, fmt.Errorf("get employee access groups error: %w", err)
	}
	if len(employeeGroups) == 0 {
		return false, nil
	}

	employeeGroupsMap := make(map[int64]struct{}, len(employeeGroups))
	for _, groupID := range employeeGroups {
		employeeGroupsMap[groupID] = struct{}{}
	}

	for _, typeAction := range ticketNotificationActionTypes {
		policyGroups := resourcePolicies[typeAction]
		if len(policyGroups) == 0 {
			continue
		}
		for _, groupID := range policyGroups {
			if _, ok := employeeGroupsMap[groupID]; ok {
				return true, nil
			}
		}
	}
	return false, nil
}

func addEmployeeIDs(target map[int64]struct{}, employeeIDs ...int64) {
	for _, employeeID := range employeeIDs {
		target[employeeID] = struct{}{}
	}
}

func (r *Realtime) processToUserChannel(ctx context.Context, employeeID int64, event models.RealtimePostEvent) error {
	publishData := models.CentrifugoUserNotifyPublishData{
		Event:    "ticket.chat.updated",
		TicketID: event.TicketID,
		Notify: models.Notify{
			Exists:        true,
			LastMessageId: event.MessageID,
			LastMessageDt: event.CreatedDt,
		},
		Meta: models.Meta{
			PublishedAt: time.Now(),
		},
	}

	channel := fmt.Sprintf("users:#%d", employeeID)

	if err := r.centrifugo.Publish(ctx, channel, publishData); err != nil {
		return fmt.Errorf("publish users event error: %w", err)
	}
	return nil
}

func (r *Realtime) processToChatNotificationChannel(ctx context.Context, notifyEmployeeIDs []int64, event models.RealtimePostEvent) error {
	publishData := models.CentrifugoChatNotificationPublishData{
		Event:              "ticket.new_chat_messages",
		TicketID:           event.TicketID,
		Message:            event.Message,
		MessageID:          event.MessageID,
		ChatID:             event.ChatID,
		SenderEmployeeID:   event.EmployeeID,
		SenderEmployeeName: event.EmployeeName,
		NotifyEmployeeIDs:  notifyEmployeeIDs,
		CreatedDt:          event.CreatedDt,
	}

	if err := r.centrifugo.Publish(ctx, "notifications:ticket_chat", publishData); err != nil {
		return fmt.Errorf("publish Chat notification event error: %w", err)
	}
	return nil
}
