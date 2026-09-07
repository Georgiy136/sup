package chat_mapper

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEmployeeID int64 = 2542

var (
	testNow                      = time.Now()
	testLastActivityAt           = testNow.Format(time.RFC3339)
	testSeenBeforeLastActivityAt = testNow.Add(-time.Minute).Format(time.RFC3339)
	testSeenAfterLastActivityAt  = testNow.Add(time.Minute).Format(time.RFC3339)
)

type redisClientMock struct {
	chatActivities map[int64]*models.ChatActivityData
	ticketIDs      []int64
	calls          int
}

func (r *redisClientMock) GetChatActivities(_ context.Context, _ int64, ticketIDs []int64) (map[int64]*models.ChatActivityData, error) {
	r.calls++
	r.ticketIDs = slices.Clone(ticketIDs)
	return r.chatActivities, nil
}

type ticketListWithChatInfoMapper interface {
	MapTicketListWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error)
}

type ticketListMapperFactory struct {
	name string
	new  func(storage ChatActivityStorage) ticketListWithChatInfoMapper
}

func ticketListMapperFactories() []ticketListMapperFactory {
	return []ticketListMapperFactory{
		{
			name: "default",
			new: func(storage ChatActivityStorage) ticketListWithChatInfoMapper {
				return New(storage)
			},
		},
		{
			name: "streaming",
			new: func(storage ChatActivityStorage) ticketListWithChatInfoMapper {
				return &ChatMapperStreaming{
					chatActivityStorage: storage,
				}
			},
		},
	}
}

type ownTicketsWithChatInfoMapper interface {
	MapOwnTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error)
}

type ownTicketsMapperFactory struct {
	name string
	new  func(storage ChatActivityStorage) ownTicketsWithChatInfoMapper
}

func ownTicketsMapperFactories() []ownTicketsMapperFactory {
	return []ownTicketsMapperFactory{
		{
			name: "default",
			new: func(storage ChatActivityStorage) ownTicketsWithChatInfoMapper {
				return New(storage)
			},
		},
	}
}

type employeeWorkTicketsWithChatInfoMapper interface {
	MapEmployeeWorkTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error)
}

type employeeWorkTicketsMapperFactory struct {
	name string
	new  func(storage ChatActivityStorage) employeeWorkTicketsWithChatInfoMapper
}

func employeeWorkTicketsMapperFactories() []employeeWorkTicketsMapperFactory {
	return []employeeWorkTicketsMapperFactory{
		{
			name: "default",
			new: func(storage ChatActivityStorage) employeeWorkTicketsWithChatInfoMapper {
				return New(storage)
			},
		},
	}
}

type createdByEmployeeTicketsWithChatInfoMapper interface {
	MapCreatedByEmployeeTicketsWithChatInfo(ctx context.Context, rawData json.RawMessage, employeeID int64) (json.RawMessage, error)
}

type createdByEmployeeTicketsMapperFactory struct {
	name string
	new  func(storage ChatActivityStorage) createdByEmployeeTicketsWithChatInfoMapper
}

func createdByEmployeeTicketsMapperFactories() []createdByEmployeeTicketsMapperFactory {
	return []createdByEmployeeTicketsMapperFactory{
		{
			name: "default",
			new: func(storage ChatActivityStorage) createdByEmployeeTicketsWithChatInfoMapper {
				return New(storage)
			},
		},
	}
}

func TestMapTicketListWithChatInfo(t *testing.T) {
	for _, factory := range ticketListMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{
					1: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					},
				},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapTicketListWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{"ticket_id": int64(1), "ticket_name": "with chat", "chat_id": "212"},
					{"ticket_id": int64(2), "ticket_name": "without chat"},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"ticket_id":   int64(1),
					"ticket_name": "with chat",
					"chat": existingChat(
						testLastActivityAt,
						testSeenBeforeLastActivityAt,
						true,
					),
				},
				{
					"ticket_id":   int64(2),
					"ticket_name": "without chat",
					"chat":        notExistingChat(),
				},
			}), string(got))
			assert.Equal(t, []int64{1}, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestMapTicketListWithChatInfoDeduplicateTicketIDs(t *testing.T) {
	for _, factory := range ticketListMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{
					1: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					},
				},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapTicketListWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{"ticket_id": int64(1), "ticket_name": "first", "chat_id": "chat_1"},
					{"ticket_id": int64(1), "ticket_name": "second", "chat_id": "chat_1"},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"ticket_id":   int64(1),
					"ticket_name": "first",
					"chat": existingChat(
						testLastActivityAt,
						testSeenBeforeLastActivityAt,
						true,
					),
				},
				{
					"ticket_id":   int64(1),
					"ticket_name": "second",
					"chat": existingChat(
						testLastActivityAt,
						testSeenBeforeLastActivityAt,
						true,
					),
				},
			}), string(got))
			assert.Equal(t, []int64{1}, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestMapTicketListWithChatInfoBatchBoundaries(t *testing.T) {
	for _, factory := range ticketListMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			const totalTickets = 1001

			chatTicketIDs := map[int64]struct{}{
				1:    {},
				500:  {},
				501:  {},
				1000: {},
				1001: {},
			}

			tickets := make([]map[string]interface{}, 0, totalTickets)
			expected := make([]map[string]interface{}, 0, totalTickets)
			chatActivities := make(map[int64]*models.ChatActivityData, len(chatTicketIDs))
			expectedTicketIDs := make([]int64, 0, len(chatTicketIDs))

			for ticketID := int64(1); ticketID <= totalTickets; ticketID++ {
				ticketName := fmt.Sprintf("ticket_%d", ticketID)

				ticket := map[string]interface{}{
					"ticket_id":   ticketID,
					"ticket_name": ticketName,
				}

				expectedTicket := map[string]interface{}{
					"ticket_id":   ticketID,
					"ticket_name": ticketName,
				}

				if _, hasChat := chatTicketIDs[ticketID]; hasChat {
					ticket["chat_id"] = fmt.Sprintf("chat_%d", ticketID)

					chatActivities[ticketID] = &models.ChatActivityData{
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					}

					expectedTicket["chat"] = existingChat(
						testLastActivityAt,
						testSeenBeforeLastActivityAt,
						true,
					)

					expectedTicketIDs = append(expectedTicketIDs, ticketID)
				} else {
					expectedTicket["chat"] = notExistingChat()
				}

				tickets = append(tickets, ticket)
				expected = append(expected, expectedTicket)
			}

			redisClient := &redisClientMock{
				chatActivities: chatActivities,
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapTicketListWithChatInfo(
				context.Background(),
				rawMessage(tickets),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON(expected), string(got))
			assert.Equal(t, expectedTicketIDs, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestMapTicketListWithChatInfoNoChats(t *testing.T) {
	for _, factory := range ticketListMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapTicketListWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{"ticket_id": int64(1), "ticket_name": "without chat 1"},
					{"ticket_id": int64(2), "ticket_name": "without chat 2"},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"ticket_id":   int64(1),
					"ticket_name": "without chat 1",
				},
				{
					"ticket_id":   int64(2),
					"ticket_name": "without chat 2",
				},
			}), string(got))
			assert.Empty(t, redisClient.ticketIDs)
		})
	}
}

func TestMapTicketListWithChatInfoError(t *testing.T) {
	for _, factory := range ticketListMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			chatMapper := factory.new(&redisClientMock{})

			got, err := chatMapper.MapTicketListWithChatInfo(
				context.Background(),
				json.RawMessage("not-json"),
				testEmployeeID,
			)

			require.Error(t, err)
			assert.Equal(t, "not-json", string(got))
		})
	}
}

func TestMapOwnTicketsWithChatInfo(t *testing.T) {
	for _, factory := range ownTicketsMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{
					10: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					},
					11: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenAfterLastActivityAt,
					},
				},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapOwnTicketsWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{
						"active_tickets": []map[string]interface{}{
							{"ticket_id": int64(10), "ticket_name": "active", "chat_id": "chat_10"},
						},
						"favorite_tickets": []map[string]interface{}{
							{"ticket_id": int64(11), "ticket_name": "favorite", "chat_id": "chat_11"},
						},
						"completed_tickets": []map[string]interface{}{
							{"ticket_id": int64(12), "ticket_name": "completed"},
						},
					},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"active_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(10),
							"ticket_name": "active",
							"chat": existingChat(
								testLastActivityAt,
								testSeenBeforeLastActivityAt,
								true,
							),
						},
					},
					"favorite_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(11),
							"ticket_name": "favorite",
							"chat": existingChat(
								testLastActivityAt,
								testSeenAfterLastActivityAt,
								false,
							),
						},
					},
					"completed_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(12),
							"ticket_name": "completed",
							"chat":        notExistingChat(),
						},
					},
					"rejected_tickets": nil,
				},
			}), string(got))
			assert.Equal(t, []int64{10, 11}, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestMapCreatedByEmployeeTicketsWithChatInfo(t *testing.T) {
	for _, factory := range createdByEmployeeTicketsMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{
					31: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					},
					32: {
						LastActivityAt: testLastActivityAt,
					},
				},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapCreatedByEmployeeTicketsWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{
						"active_tickets": []map[string]interface{}{
							{"ticket_id": int64(31), "ticket_name": "active", "chat_id": "chat_31"},
						},
						"rejected_tickets": []map[string]interface{}{
							{"ticket_id": int64(32), "ticket_name": "rejected", "chat_id": "chat_32"},
						},
						"completed_tickets": []map[string]interface{}{
							{"ticket_id": int64(33), "ticket_name": "completed"},
						},
					},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"active_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(31),
							"ticket_name": "active",
							"chat": existingChat(
								testLastActivityAt,
								testSeenBeforeLastActivityAt,
								true,
							),
						},
					},
					"rejected_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(32),
							"ticket_name": "rejected",
							"chat": existingChat(
								testLastActivityAt,
								nil,
								true,
							),
						},
					},
					"completed_tickets": []map[string]interface{}{
						{
							"ticket_id":   int64(33),
							"ticket_name": "completed",
							"chat":        notExistingChat(),
						},
					},
				},
			}), string(got))
			assert.Equal(t, []int64{31, 32}, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestMapEmployeeWorkTicketsWithChatInfo(t *testing.T) {
	for _, factory := range employeeWorkTicketsMapperFactories() {
		t.Run(factory.name, func(t *testing.T) {
			redisClient := &redisClientMock{
				chatActivities: map[int64]*models.ChatActivityData{
					21: {
						LastActivityAt:     testLastActivityAt,
						LastUserActivityAt: testSeenBeforeLastActivityAt,
					},
					22: {
						LastActivityAt: testLastActivityAt,
					},
				},
			}
			chatMapper := factory.new(redisClient)

			got, err := chatMapper.MapEmployeeWorkTicketsWithChatInfo(
				context.Background(),
				rawMessage([]map[string]interface{}{
					{
						"approve": []map[string]interface{}{
							{"ticket_id": int64(21), "ticket_name": "approve", "chat_id": "chat_21"},
						},
						"perform": []map[string]interface{}{
							{"ticket_id": int64(22), "ticket_name": "perform", "chat_id": "chat_22"},
						},
						"booked_perform": []map[string]interface{}{
							{"ticket_id": int64(23), "ticket_name": "booked"},
						},
					},
				}),
				testEmployeeID,
			)

			require.NoError(t, err)
			assert.JSONEq(t, rawJSON([]map[string]interface{}{
				{
					"approve": []map[string]interface{}{
						{
							"ticket_id":   int64(21),
							"ticket_name": "approve",
							"chat": existingChat(
								testLastActivityAt,
								testSeenBeforeLastActivityAt,
								true,
							),
						},
					},
					"perform": []map[string]interface{}{
						{
							"ticket_id":   int64(22),
							"ticket_name": "perform",
							"chat": existingChat(
								testLastActivityAt,
								nil,
								true,
							),
						},
					},
					"booked_perform": []map[string]interface{}{
						{
							"ticket_id":   int64(23),
							"ticket_name": "booked",
							"chat":        notExistingChat(),
						},
					},
				},
			}), string(got))
			assert.Equal(t, []int64{21, 22}, redisClient.ticketIDs)
			assert.Equal(t, 1, redisClient.calls)
		})
	}
}

func TestWalkTicketBatches(t *testing.T) {
	rawData := rawMessage([]map[string]interface{}{
		{"ticket_id": int64(1), "ticket_name": "ticket 1", "chat_id": "chat_1"},
		{"ticket_id": int64(2), "ticket_name": "ticket 2", "chat_id": "chat_2"},
		{"ticket_id": int64(3), "ticket_name": "ticket 3", "chat_id": "chat_3"},
		{"ticket_id": int64(4), "ticket_name": "ticket 4", "chat_id": "chat_4"},
		{"ticket_id": int64(5), "ticket_name": "ticket 5", "chat_id": "chat_5"},
	})

	var batchSizes []int
	var gotTicketIDs []int64

	err := walkTicketBatches(context.Background(), rawData, 2, func(batch []ticketWithChat) error {
		batchSizes = append(batchSizes, len(batch))

		for i := range batch {
			ticketID, hasChat := batch[i].getTicketIDAndChatExists()
			require.True(t, hasChat)
			gotTicketIDs = append(gotTicketIDs, ticketID)
		}

		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []int{2, 2, 1}, batchSizes)
	assert.Equal(t, []int64{1, 2, 3, 4, 5}, gotTicketIDs)
}

func rawMessage(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func existingChat(lastActivityAt string, lastUserActivityAt interface{}, hasUnread bool) map[string]interface{} {
	return map[string]interface{}{
		"exists":                true,
		"has_unread":            hasUnread,
		"last_activity_at":      lastActivityAt,
		"last_user_activity_at": lastUserActivityAt,
	}
}

func notExistingChat() map[string]interface{} {
	return map[string]interface{}{
		"exists":                false,
		"has_unread":            false,
		"last_activity_at":      nil,
		"last_user_activity_at": nil,
	}
}

func rawJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
