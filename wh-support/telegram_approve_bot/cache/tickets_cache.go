package cache

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type TicketsCache struct {
	cache *redis.Client
	ttl   time.Duration
}

func NewTicketsCache(ctx context.Context, config configs.Config) *TicketsCache {
	redisCfgKey := "tickets_cache_config"
	redisConfBytes := config.GetByServiceKeyRequired(redisCfgKey)

	const (
		modeSentinel   = "sentinel"
		modeStandalone = "standalone"
	)

	var redisClient *redis.Client

	var redisConfiguration struct {
		MasterName       string   `json:"MasterName"`
		SentinelAddrs    []string `json:"SentinelAddrs"`
		Addr             string   `json:"Addr"`
		SentinelPassword string   `json:"SentinelPassword"`
		Password         string   `json:"Password"`
		TTL              string   `json:"TTL"`
		Mode             string   `json:"Mode"`
	}
	err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration)
	if err != nil {
		logrus.Panicf("error unmarshaling redis config: %v", err)
	}

	ttl, err := time.ParseDuration(redisConfiguration.TTL)
	if err != nil {
		logrus.Panicf("error parsing redis TTL: %v", err)
	}

	switch redisConfiguration.Mode {
	case modeStandalone:
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfiguration.Addr,
			Password: redisConfiguration.Password,
		})
	case modeSentinel:
		redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConfiguration.MasterName,
			SentinelAddrs:    redisConfiguration.SentinelAddrs,
			SentinelPassword: redisConfiguration.SentinelPassword,
			Password:         redisConfiguration.Password,
		})
	default:
		logrus.Panicf("unknown redis mode '%s'", redisConfiguration.Mode)
	}

	if err = redisClient.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, check redis_configuration.json, err: %v", err)
	}
	logrus.Debugf("redis client inited")
	return &TicketsCache{
		cache: redisClient,
		ttl:   ttl,
	}
}

var (
	ticketPrefix    = "ticket%d_chatid_%d"
	ticketPrefixOld = "ticket_%d"
)

func (c *TicketsCache) GetTicketsByChatID(ticketID, chatID int64) ([]models.TicketData, error) {
	results, err := c.cache.LRange(context.Background(), fmt.Sprintf(ticketPrefix, ticketID, chatID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("can't get tickets from cache: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	msgs := make([]models.TicketData, 0, len(results))
	for _, arg := range results {
		var msg models.TicketData
		if err = jsoniter.Unmarshal([]byte(arg), &msg); err != nil {
			return nil, fmt.Errorf("can't get tickets, unmarshal msg to json error: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (c *TicketsCache) Delete(ticketID, chatID int64) error {
	if err := c.cache.Del(context.Background(), fmt.Sprintf(ticketPrefix, ticketID, chatID)).Err(); err != nil {
		return fmt.Errorf("can't delete ticket from cache: %w", err)
	}
	return nil
}

func (c *TicketsCache) DeleteTicketWithOldPrefix(ticketID int64) error {
	if err := c.cache.Del(context.Background(), fmt.Sprintf(ticketPrefixOld, ticketID)).Err(); err != nil {
		return fmt.Errorf("can't delete ticket from cache: %w", err)
	}
	return nil
}

func (c *TicketsCache) AppendTicket(ticketID int64, ticketData models.TicketData) error {
	data, err := jsoniter.MarshalToString(ticketData)
	if err != nil {
		return fmt.Errorf("can't marshal msg to json: %v", err)
	}
	if err = c.cache.LPush(context.Background(), fmt.Sprintf(ticketPrefix, ticketID, ticketData.Message.ChatID), data).Err(); err != nil {
		return fmt.Errorf("can't append ticket to cache: %w", err)
	}
	c.cache.Expire(context.Background(), fmt.Sprintf(ticketPrefix, ticketID, ticketData.Message.ChatID), c.ttl)
	return nil
}

func (c *TicketsCache) GetTicketsWithOldPrefix(ticketID int64) ([]models.TicketData, error) {
	results, err := c.cache.LRange(context.Background(), fmt.Sprintf(ticketPrefixOld, ticketID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("can't get tickets from cache: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	msgs := make([]models.TicketData, 0, len(results))
	for _, arg := range results {
		var msg models.TicketData
		if err = jsoniter.Unmarshal([]byte(arg), &msg); err != nil {
			return nil, fmt.Errorf("can't get tickets, unmarshal msg to json error: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (c *TicketsCache) DeleteByOperationType(ticketID, chatID int64, operationType string) error {
	ticketMsgs, err := c.GetTicketsByChatID(ticketID, chatID)
	if err != nil {
		return fmt.Errorf("can not get ticket %d: %w", ticketID, err)
	}

	stableTicketMsgs := make([]models.TicketData, 0)
	for _, ticketMsg := range ticketMsgs {
		if ticketMsg.Operation != operationType {
			stableTicketMsgs = append(stableTicketMsgs, ticketMsg)
		}
	}

	switch {
	case len(stableTicketMsgs) == 0:
		err = c.Delete(ticketID, chatID)
		if err != nil {
			return fmt.Errorf("can't delete tickets: %v", err)
		}
	case len(stableTicketMsgs) < len(ticketMsgs):
		err := c.Delete(ticketID, chatID)
		if err != nil {
			return fmt.Errorf("can't delete tickets: %v", err)
		}

		for _, msg := range stableTicketMsgs {
			err := c.AppendTicket(ticketID, msg)
			if err != nil {
				logrus.Errorf("can't append stable ticket: %v", err)
			}
		}
	}

	return nil
}
