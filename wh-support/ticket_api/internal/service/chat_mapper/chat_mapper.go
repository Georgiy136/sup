package chat_mapper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type ChatMapper struct {
	chatActivityStorage ChatActivityStorage
}

func New(chatActivityStorage ChatActivityStorage) *ChatMapper {
	return &ChatMapper{chatActivityStorage: chatActivityStorage}
}

type ChatActivityStorage interface {
	GetChatActivities(ctx context.Context, employeeID int64, ticketIDs []int64) (map[int64]*models.ChatActivityData, error)
}

func (c *ChatMapper) MapTicketListWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error) {
	const op = "MapTicketListWithChatInfo"

	utils.Log("01_before_unmarshal", employeeID, logrus.Fields{
		"emp_d":        employeeID,
		"raw_size_mib": len(rawData) / 1024 / 1024,
	})

	var tickets []ticketWithChat
	if err := json.Unmarshal(rawData, &tickets); err != nil {
		return rawData, fmt.Errorf("%s: can't unmarshal tickets: %w", op, err)
	}

	utils.Log("02_after_unmarshal", employeeID, logrus.Fields{
		"employee_id": employeeID,
		"tickets_len": len(tickets),
		"tickets_cap": cap(tickets),
	})

	if err := c.fillTicketsChatInfo(ctx, employeeID, tickets); err != nil {
		return rawData, fmt.Errorf("%s: can't fill chat info: %w", op, err)
	}

	utils.Log("07_after_fill_chat_info", employeeID, logrus.Fields{
		"employee_id": employeeID,
		"tickets_len": len(tickets),
		"tickets_cap": cap(tickets),
	})

	rawResult, err := jsoniter.Marshal(tickets)
	if err != nil {
		return nil, fmt.Errorf("%s: can't marshal tickets: %w", op, err)
	}

	utils.Log("08_after_marshal", employeeID, logrus.Fields{
		"employee_id":     employeeID,
		"result_size_mib": len(rawResult) / 1024 / 1024,
	})

	return rawResult, nil
}

func (c *ChatMapper) MapCreatedByEmployeeTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error) {
	const op = "MapCreatedByEmployeeTicketsWithChatInfo"

	var groups []createdByEmployeeTickets
	if err := jsoniter.Unmarshal(rawData, &groups); err != nil {
		return rawData, fmt.Errorf("%s: can't unmarshal tickets: %w", op, err)
	}

	ticketLists := make([][]ticketWithChat, 0, len(groups)*3)
	for i := range groups {
		ticketLists = append(ticketLists,
			groups[i].ActiveTickets,
			groups[i].RejectedTickets,
			groups[i].CompletedTickets,
		)
	}
	if err := c.fillTicketsChatInfo(ctx, employeeID, ticketLists...); err != nil {
		return rawData, fmt.Errorf("%s: can't fill chat info: %w", op, err)
	}

	rawResult, err := jsoniter.Marshal(groups)
	if err != nil {
		return nil, fmt.Errorf("%s: can't marshal tickets: %w", op, err)
	}

	return rawResult, nil
}

func (c *ChatMapper) MapOwnTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error) {
	const op = "MapOwnTicketsWithChatInfo"

	var groups []ownTickets
	if err := jsoniter.Unmarshal(rawData, &groups); err != nil {
		return rawData, fmt.Errorf("%s: can't unmarshal tickets: %w", op, err)
	}

	ticketLists := make([][]ticketWithChat, 0, len(groups)*4)
	for i := range groups {
		ticketLists = append(ticketLists,
			groups[i].ActiveTickets,
			groups[i].FavoriteTickets,
			groups[i].RejectedTickets,
			groups[i].CompletedTickets,
		)
	}
	if err := c.fillTicketsChatInfo(ctx, employeeID, ticketLists...); err != nil {
		return rawData, fmt.Errorf("%s: can't fill chat info: %w", op, err)
	}
	rawResult, err := jsoniter.Marshal(groups)
	if err != nil {
		return nil, fmt.Errorf("%s: can't marshal tickets: %w", op, err)
	}

	return rawResult, nil
}

func (c *ChatMapper) MapEmployeeWorkTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error) {
	const op = "MapEmployeeWorkTicketsWithChatInfo"

	var groups []employeeWorkTickets
	if err := jsoniter.Unmarshal(rawData, &groups); err != nil {
		return rawData, fmt.Errorf("%s: can't unmarshal tickets: %w", op, err)
	}

	ticketLists := make([][]ticketWithChat, 0, len(groups)*3)
	for i := range groups {
		ticketLists = append(ticketLists,
			groups[i].Approve,
			groups[i].Perform,
			groups[i].BookedPerform,
		)
	}
	if err := c.fillTicketsChatInfo(ctx, employeeID, ticketLists...); err != nil {
		return rawData, fmt.Errorf("%s: can't fill chat info: %w", op, err)
	}

	rawResult, err := jsoniter.Marshal(groups)
	if err != nil {
		return nil, fmt.Errorf("%s: can't marshal tickets: %w", op, err)
	}

	return rawResult, nil
}

func (c *ChatMapper) fillTicketsChatInfo(ctx context.Context, employeeID int64, ticketLists ...[]ticketWithChat) error {
	utils.Log("03_before_collect_ticket_ids", employeeID, logrus.Fields{
		"employee_id": employeeID,
	})

	ticketIDs := collectTicketsWithChatID(ticketLists...)
	if len(ticketIDs) == 0 {
		return nil
	}

	utils.Log("04_after_collect_ticket_ids", employeeID, logrus.Fields{
		"employee_id":    employeeID,
		"ticket_ids_len": len(ticketIDs),
		"ticket_ids_cap": cap(ticketIDs),
	})

	activities, err := c.chatActivityStorage.GetChatActivities(ctx, employeeID, ticketIDs)
	if err != nil {
		return fmt.Errorf("can't get chat activities: %w", err)
	}

	utils.Log("05_after_get_chat_activities", employeeID, logrus.Fields{
		"employee_id":    employeeID,
		"activities_len": len(activities),
	})

	for _, sectionTickets := range ticketLists {
		for i := range sectionTickets {
			ticketID, hasChat := sectionTickets[i].getTicketIDAndChatExists()
			if !hasChat {
				if err = sectionTickets[i].setChat(nil); err != nil {
					return err
				}
				continue
			}
			if err = sectionTickets[i].setChat(buildChatInfo(activities[ticketID])); err != nil {
				return err
			}
		}
	}

	utils.Log("06_after_set_chat_info", employeeID, logrus.Fields{
		"employee_id": employeeID,
	})

	return nil
}

func collectTicketsWithChatID(ticketLists ...[]ticketWithChat) []int64 {
	totalTickets := 0
	for _, sectionTickets := range ticketLists {
		totalTickets += len(sectionTickets)
	}

	ticketsWithChatMap := make(map[int64]struct{}, totalTickets)
	ticketsIDs := make([]int64, 0, totalTickets)

	for _, sectionTickets := range ticketLists {
		for _, ticket := range sectionTickets {
			ticketID, hasChat := ticket.getTicketIDAndChatExists()
			if !hasChat {
				continue
			}
			if _, ok := ticketsWithChatMap[ticketID]; ok {
				continue
			}
			ticketsWithChatMap[ticketID] = struct{}{}
			ticketsIDs = append(ticketsIDs, ticketID)
		}
	}

	return ticketsIDs
}

func buildChatInfo(activity *models.ChatActivityData) *models.ChatInfo {
	chatInfo := &models.ChatInfo{}

	if activity == nil {
		return chatInfo
	}
	if activity.LastActivityAt != "" {
		chatInfo.LastActivityAt = &activity.LastActivityAt
	}
	if activity.LastUserActivityAt != "" {
		chatInfo.LastUserActivityAt = &activity.LastUserActivityAt
	}

	chatInfo.HasUnread = hasUnread(activity)
	chatInfo.Exists = true
	return chatInfo
}

func hasUnread(activity *models.ChatActivityData) bool {
	if activity.LastActivityAt == "" {
		return false
	}
	if activity.LastUserActivityAt == "" {
		return true
	}

	lastActivity, err := parseChatTime(activity.LastActivityAt)
	if err != nil {
		logrus.Errorf("can't parse last_activity_at '%s': %v", activity.LastActivityAt, err)
		return false
	}

	lastSeen, err := parseChatTime(activity.LastUserActivityAt)
	if err != nil {
		logrus.Errorf("can't parse last_user_activity_at '%s': %v", activity.LastUserActivityAt, err)
		return false
	}

	return lastActivity.After(lastSeen)
}

func parseChatTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}
