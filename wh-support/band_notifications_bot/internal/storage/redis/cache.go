package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	chatmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/chat/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	*redis.Client
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Configure(ctx context.Context, config configs.Config) {
	redisConfBytes := config.GetByServiceKeyRequired("redis_configuration")

	const (
		modeSentinel   = "sentinel"
		modeStandalone = "standalone"
	)

	var redisConfiguration struct {
		MasterName       string   `json:"MasterName"`
		SentinelAddrs    []string `json:"SentinelAddrs"`
		Addr             string   `json:"Addr"`
		SentinelPassword string   `json:"SentinelPassword"`
		Password         string   `json:"Password"` //nolint:gosec
		Mode             string   `json:"Mode"`
	}
	if err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration); err != nil {
		logrus.Panicf("error unmarshaling redis config: %v", err)
	}

	switch redisConfiguration.Mode {
	case modeStandalone:
		s.Client = redis.NewClient(&redis.Options{
			Addr:     redisConfiguration.Addr,
			Password: redisConfiguration.Password,
		})
	case modeSentinel:
		s.Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConfiguration.MasterName,
			SentinelAddrs:    redisConfiguration.SentinelAddrs,
			SentinelPassword: redisConfiguration.SentinelPassword,
			Password:         redisConfiguration.Password,
		})
	default:
		logrus.Panicf("unknown redis mode '%s'", redisConfiguration.Mode)
	}

	if err := s.Client.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, err: %v", err)
	}
}

const (
	redisKeyBand                = "chat_notifications:last_offset"
	lastChatViewByUserKeyFormat = "chat_user:%d:ticket_id:%d:last_seen"
)

func (s *Storage) GetSavedOffset(ctx context.Context) (*models.HistoryOffset, error) {
	raw, err := s.Client.Get(ctx, redisKeyBand).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, customerrors.ErrKeyNotFound
		}
		return nil, fmt.Errorf("redis get band cursor: %w", err)
	}

	var cursor models.HistoryOffset
	if err = jsoniter.UnmarshalFromString(raw, &cursor); err != nil {
		return nil, fmt.Errorf("unmarshal band cursor: %w", err)
	}
	return &cursor, nil
}

func (s *Storage) SaveOffset(ctx context.Context, pos models.HistoryOffset) error {
	raw, err := jsoniter.MarshalToString(pos)
	if err != nil {
		return fmt.Errorf("marshal band cursor: %w", err)
	}

	if err = s.Client.Set(ctx, redisKeyBand, raw, 0).Err(); err != nil {
		return fmt.Errorf("redis set band cursor: %w", err)
	}

	return nil
}

func (s *Storage) GetLastChatViewsByUsers(ctx context.Context, requests []chatmodels.LastChatViewRequest) (map[chatmodels.LastChatViewRequest]time.Time, error) {
	if len(requests) == 0 {
		return map[chatmodels.LastChatViewRequest]time.Time{}, nil
	}

	pipe := s.Client.Pipeline()
	cmds := make(map[chatmodels.LastChatViewRequest]*redis.StringCmd, len(requests))
	for _, request := range requests {
		if _, ok := cmds[request]; ok {
			continue
		}
		cmds[request] = pipe.Get(ctx, getLastChatViewByUserKey(request.EmployeeID, request.TicketID))
	}

	if len(cmds) == 0 {
		return map[chatmodels.LastChatViewRequest]time.Time{}, nil
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("can't get last chat views from redis: %w", err)
	}

	result := make(map[chatmodels.LastChatViewRequest]time.Time, len(cmds))
	for request, cmd := range cmds {
		raw, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return nil, fmt.Errorf("can't get last chat view from redis: %w", err)
		}

		viewedAt, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			logrus.Warnf("can't parse last chat view time %q for employee %d ticket %d: %v", raw, request.EmployeeID, request.TicketID, err)
			continue
		}
		result[request] = viewedAt
	}

	return result, nil
}

func getLastChatViewByUserKey(employeeID, ticketID int64) string {
	return fmt.Sprintf(lastChatViewByUserKeyFormat, employeeID, ticketID)
}
