package redis

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/consts"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"strconv"
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

func (r *RedisClient) GetAccessData(ctx context.Context, typeAction string, employeeID int64) (*models.AccessData, error) {
	pipe := r.Pipeline()

	externalActionsCmd := pipe.SMembers(ctx, getExternalActionsKey(typeAction))
	actionGroupsCmd := pipe.SMembers(ctx, getActionGroupKey(typeAction))
	employeeGroupsCmd := pipe.SMembers(ctx, getEmployeeActionKey(employeeID))

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("can't execute pipeline: %w", err)
	}

	externalActions, err := r.parseExternalActions(externalActionsCmd)
	if err != nil {
		return nil, err
	}

	actionGroups, err := r.parseActionGroups(actionGroupsCmd)
	if err != nil {
		return nil, err
	}

	employeeGroups, err := r.parseEmployeeGroups(employeeGroupsCmd)
	if err != nil {
		return nil, err
	}

	return &models.AccessData{
		ExternalActions: externalActions,
		ActionGroups:    actionGroups,
		EmployeeGroups:  employeeGroups,
	}, nil
}

func (r *RedisClient) parseExternalActions(cmd *redis.StringSliceCmd) ([]models.ExternalActionWithCategory, error) {
	result, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get external actions from redis, err: %w", err)
	}
	if len(result) == 0 {
		return nil, nil
	}

	externalActions := make([]models.ExternalActionWithCategory, len(result))
	for i := range result {
		if err = jsoniter.Unmarshal([]byte(result[i]), &externalActions[i]); err != nil {
			return nil, fmt.Errorf("can't unmarshal external action from redis, err: %w", err)
		}
	}

	return externalActions, nil
}

func (r *RedisClient) parseActionGroups(cmd *redis.StringSliceCmd) ([]models.ActionGroupInfoWithCategory, error) {
	result, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get action groups from redis, err: %w", err)
	}
	if len(result) == 0 {
		return nil, nil
	}

	actionGroups := make([]models.ActionGroupInfoWithCategory, len(result))
	for i := range result {
		if err = jsoniter.Unmarshal([]byte(result[i]), &actionGroups[i]); err != nil {
			return nil, fmt.Errorf("can't unmarshal action group from redis, err: %w", err)
		}
	}

	return actionGroups, nil
}

func (r *RedisClient) parseEmployeeGroups(cmd *redis.StringSliceCmd) ([]int64, error) {
	result, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get employee groups from redis, err: %w", err)
	}
	if len(result) == 0 {
		return nil, nil
	}

	employeeGroups := make([]int64, len(result))
	for i := range result {
		groupID, err := strconv.ParseInt(result[i], consts.Base10, consts.BitSize64)
		if err != nil {
			return nil, fmt.Errorf("can't parse employee groupID from redis, err: %w", err)
		}
		employeeGroups[i] = groupID
	}

	return employeeGroups, nil
}
