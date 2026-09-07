package clients

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(ctx context.Context, config configs.Config) *RedisClient {
	redisCfgKey := "redis_configuration"
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
		Password         string   `json:"Password"` //nolint:gosec
		Mode             string   `json:"Mode"`
	}
	err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration)
	if err != nil {
		logrus.Panicf("error unmarshaling redis config: %v", err)
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
	return &RedisClient{redisClient}
}

// GetFromCache вернет (nil, nil), если в редисе нет информации по employeeID
func (c *RedisClient) GetFromCache(ctx context.Context, employeeID int64) (*models.EmployeeResources, error) {
	resString, err := c.Get(ctx, strconv.FormatInt(employeeID, 16)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get data from redis, err: %w", err)
	}
	result := new(models.EmployeeResources)
	err = jsoniter.Unmarshal([]byte(resString), result)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data from cache, err: %w", err)
	}
	return result, nil
}

func (c *RedisClient) GetFromCacheByExternalID(ctx context.Context, employeeID int64) ([]string, error) {
	resString, err := c.Get(ctx, getKeyFromEmployeeID(employeeID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get data from redis, err: %w", err)
	}
	result := make([]string, 0)
	err = jsoniter.Unmarshal([]byte(resString), &result)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data from cache, err: %w", err)
	}
	return result, nil
}

func (c *RedisClient) SaveToCacheByExternalID(ctx context.Context, employeeID int64, data []string) error {
	bytes, err := jsoniter.Marshal(data)
	if err != nil {
		return fmt.Errorf("can't marshal data, err: %w", err)
	}
	err = c.Set(ctx, getKeyFromEmployeeID(employeeID), bytes, 1*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("can't save data to redis, err: %w", err)
	}
	return nil
}

func getKeyFromEmployeeID(employeeID int64) string {
	id := strconv.FormatInt(employeeID, 16)
	return "external_" + id
}
