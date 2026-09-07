package redis

import (
	"context"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	rCli *redis.Client
}

func NewRedisClient() *RedisClient {
	return &RedisClient{new(redis.Client)}
}

func (rc *RedisClient) Configure(ctx context.Context, cfg configs.Config) {
	redisCfgKey := "redis_configuration"
	redisConfBytes := cfg.GetByServiceKeyRequired(redisCfgKey)

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
	err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration)
	if err != nil {
		logrus.Panicf("error unmarshaling redis config: %v", err)
	}

	switch redisConfiguration.Mode {
	case modeStandalone:
		rc.rCli = redis.NewClient(&redis.Options{
			Addr:     redisConfiguration.Addr,
			Password: redisConfiguration.Password,
		})
	case modeSentinel:
		rc.rCli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConfiguration.MasterName,
			SentinelAddrs:    redisConfiguration.SentinelAddrs,
			SentinelPassword: redisConfiguration.SentinelPassword,
			Password:         redisConfiguration.Password,
		})
	default:
		logrus.Panicf("unknown redis mode '%s'", redisConfiguration.Mode)
	}

	if err = rc.rCli.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, check redis_configuration.json, err: %v", err)
	}
	logrus.Debugf("redis client inited")
}
