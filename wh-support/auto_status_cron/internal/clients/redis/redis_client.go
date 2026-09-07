package redis

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient() *RedisClient {
	return &RedisClient{}
}

func (r *RedisClient) Configure(ctx context.Context, config configs.Config) {
	redisCfgKey := "redis_configuration"
	redisConfBytes := config.GetByServiceKeyRequired(redisCfgKey)

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
		r.Client = redis.NewClient(&redis.Options{
			Addr:     redisConfiguration.Addr,
			Password: redisConfiguration.Password,
		})
	case modeSentinel:
		r.Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConfiguration.MasterName,
			SentinelAddrs:    redisConfiguration.SentinelAddrs,
			SentinelPassword: redisConfiguration.SentinelPassword,
			Password:         redisConfiguration.Password,
		})
	default:
		logrus.Panicf("unknown redis mode '%s'", redisConfiguration.Mode)
	}

	if err = r.Client.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, check redis_configuration.json, err: %v", err)
	}
	logrus.Debugf("redis client inited")
}

func (r *RedisClient) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) (err error) {
	const redisOp = "SET"
	span := sentry.StartDBRedisSSpan(ctx, redisOp, key)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Del(ctx context.Context, key string) (err error) {
	const redisOp = "DEL"
	span := sentry.StartDBRedisSSpan(ctx, redisOp, key)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (exist bool, err error) {
	const redisOp = "EXISTS"
	span := sentry.StartDBRedisSSpan(ctx, redisOp, key)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()
	n, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check redis exists error: %w", err)
	}
	return n > 0, nil
}
