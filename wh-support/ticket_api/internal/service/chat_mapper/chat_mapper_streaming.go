package chat_mapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type ChatMapperStreaming struct {
	chatActivityStorage ChatActivityStorage
}

func NewChatMapperStreaming(chatActivityStorage ChatActivityStorage) *ChatMapperStreaming {
	return &ChatMapperStreaming{chatActivityStorage}
}

func (c *ChatMapperStreaming) MapTicketListWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error) {
	const op = "MapTicketListWithChatInfo"

	const batchSize = 500

	utils.Log("01_before_stream_collect_ticket_ids", employeeID, logrus.Fields{
		"employee_id":  employeeID,
		"raw_size_mib": len(rawData) / 1024 / 1024,
		"batch_size":   batchSize,
	})

	ticketIDs, err := collectTicketsWithChatIDFromRawBatched(ctx, rawData, batchSize)
	if err != nil {
		return rawData, fmt.Errorf("%s: can't collect ticket ids: %w", op, err)
	}

	utils.Log("02_after_stream_collect_ticket_ids", employeeID, logrus.Fields{
		"employee_id":    employeeID,
		"ticket_ids_len": len(ticketIDs),
		"ticket_ids_cap": cap(ticketIDs),
	})

	var activities map[int64]*models.ChatActivityData
	if len(ticketIDs) > 0 {
		activities, err = c.chatActivityStorage.GetChatActivities(ctx, employeeID, ticketIDs)
		if err != nil {
			return rawData, fmt.Errorf("%s: can't get chat activities: %w", op, err)
		}
	}

	utils.Log("06_after_get_chat_activities", employeeID, logrus.Fields{
		"employee_id":    employeeID,
		"activities_len": len(activities),
	})

	rawResult, err := buildTicketListWithChatInfoRawBatched(ctx, rawData, activities, batchSize)
	if err != nil {
		return rawData, fmt.Errorf("%s: can't build result: %w", op, err)
	}

	utils.Log("07_after_batch_build_result", employeeID, logrus.Fields{
		"employee_id":     employeeID,
		"result_size_mib": len(rawResult) / 1024 / 1024,
		"batch_size":      batchSize,
	})

	return rawResult, nil
}

func collectTicketsWithChatIDFromRawBatched(ctx context.Context, rawData json.RawMessage, batchSize int64) ([]int64, error) {
	seen := make(map[int64]struct{})
	ticketIDs := make([]int64, 0)

	err := walkTicketBatches(ctx, rawData, batchSize, func(batch []ticketWithChat) error {
		for i := range batch {
			ticketID, hasChat := batch[i].getTicketIDAndChatExists()
			if !hasChat {
				continue
			}

			if _, ok := seen[ticketID]; ok {
				continue
			}

			seen[ticketID] = struct{}{}
			ticketIDs = append(ticketIDs, ticketID)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't walk ticket batch: %w", err)
	}

	return ticketIDs, nil
}

func buildTicketListWithChatInfoRawBatched(ctx context.Context, rawData json.RawMessage, activities map[int64]*models.ChatActivityData, batchSize int64) (json.RawMessage, error) {
	var buf bytes.Buffer

	buf.Grow(len(rawData) + len(rawData)/10)

	buf.WriteByte('[')

	firstTicket := true

	err := walkTicketBatches(ctx, rawData, batchSize, func(batch []ticketWithChat) error {
		if len(activities) != 0 {
			if err := fillChatInfoByTicketBatchWithActivities(batch, activities); err != nil {
				return fmt.Errorf("can't fill batch: %w", err)
			}
		}

		for i := range batch {
			if !firstTicket {
				buf.WriteByte(',')
			}
			firstTicket = false

			ticketRaw, err := jsoniter.Marshal(batch[i])
			if err != nil {
				return fmt.Errorf("can't marshal ticket: %w", err)
			}

			buf.Write(ticketRaw)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't walk ticket batches: %w", err)
	}

	buf.WriteByte(']')

	return buf.Bytes(), nil
}

func fillChatInfoByTicketBatchWithActivities(tickets []ticketWithChat, activities map[int64]*models.ChatActivityData) error {
	for i := range tickets {
		ticketID, hasChat := tickets[i].getTicketIDAndChatExists()
		if !hasChat {
			if err := tickets[i].setChat(nil); err != nil {
				return fmt.Errorf("can't set chat without activity ticket %d: %w", ticketID, err)
			}
			continue
		}

		if err := tickets[i].setChat(buildChatInfo(activities[ticketID])); err != nil {
			return fmt.Errorf("can't set chat with activity ticket %d: %w", ticketID, err)
		}
	}

	return nil
}

type TicketBatchHandler func(batch []ticketWithChat) error

func walkTicketBatches(ctx context.Context, rawData json.RawMessage, batchSize int64, handle TicketBatchHandler) error {
	if batchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}

	dec := json.NewDecoder(bytes.NewReader(rawData))

	if err := expectDelim(dec, '['); err != nil {
		return fmt.Errorf("can't expect delim: %w", err)
	}

	batch := make([]ticketWithChat, 0, batchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error: %w", err)
		}

		if err := handle(batch); err != nil {
			return fmt.Errorf("handle error: %w", err)
		}

		for i := range batch {
			batch[i] = ticketWithChat{}
		}

		batch = batch[:0]

		return nil
	}

	for dec.More() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error: %w", err)
		}

		var ticket ticketWithChat
		if err := dec.Decode(&ticket); err != nil {
			return fmt.Errorf("can't decode ticket: %w", err)
		}

		batch = append(batch, ticket)

		if int64(len(batch)) == batchSize {
			if err := flush(); err != nil {
				return fmt.Errorf("can't flush ticket batches: %w", err)
			}
		}
	}

	if err := flush(); err != nil {
		return fmt.Errorf("can't flush ticket batches: %w", err)
	}

	if err := expectDelim(dec, ']'); err != nil {
		return fmt.Errorf("can't expect delim: %w", err)
	}

	return nil
}

func expectDelim(dec *json.Decoder, expected json.Delim) error {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("can't read token: %w", err)
	}

	delim, ok := token.(json.Delim)
	if !ok || delim != expected {
		return fmt.Errorf("expected delimiter %q, got %v", expected, token)
	}

	return nil
}
