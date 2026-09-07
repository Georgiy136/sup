package redis

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const redisConfigurationKey = "redis_configuration"

type Cache struct {
	Client *redis.Client
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Configure(ctx context.Context, config configs.Config) {
	redisConfBytes := config.GetByServiceKeyRequired(redisConfigurationKey)

	const (
		modeSentinel   = "sentinel"
		modeStandalone = "standalone"
	)

	//nolint:tagliatelle
	var redisConfiguration struct {
		MasterName       string   `json:"MasterName"`
		SentinelAddrs    []string `json:"SentinelAddrs"`
		Addr             string   `json:"Addr"`
		SentinelPassword string   `json:"SentinelPassword"`
		Password         string   `json:"Password"`
		Mode             string   `json:"Mode"`
	}
	if err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration); err != nil {
		logrus.Panicf("error unmarshaling redis config: %v", err)
	}

	switch redisConfiguration.Mode {
	case modeStandalone:
		c.Client = redis.NewClient(&redis.Options{
			Addr:     redisConfiguration.Addr,
			Password: redisConfiguration.Password,
		})
	case modeSentinel:
		c.Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConfiguration.MasterName,
			SentinelAddrs:    redisConfiguration.SentinelAddrs,
			SentinelPassword: redisConfiguration.SentinelPassword,
			Password:         redisConfiguration.Password,
		})
	default:
		logrus.Panicf("unknown redis mode '%s'", redisConfiguration.Mode)
	}

	if err := c.Client.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, err: %v", err)
	}
}
