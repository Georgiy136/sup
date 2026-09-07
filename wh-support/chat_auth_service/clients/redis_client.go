package clients

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/models"
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
	if err := jsoniter.Unmarshal(redisConfBytes, &redisConfiguration); err != nil {
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

	if err := r.Client.Ping(ctx).Err(); err != nil {
		logrus.Panicf("can't create redis client, check redis_configuration.json, err: %v", err)
	}
	logrus.Debugf("redis client inited")
}

func (r *RedisClient) GetAccessPolicies(ctx context.Context, employeeID int64, typeActions ...string) (*models.AccessPoliciesResult, error) {
	pipe := r.Client.Pipeline()

	type actionCommands struct {
		externalActionsCmd *redis.StringSliceCmd
		actionGroupsCmd    *redis.StringSliceCmd
	}

	commandsMap := make(map[string]*actionCommands, len(typeActions))
	for _, typeAction := range typeActions {
		commandsMap[typeAction] = &actionCommands{
			externalActionsCmd: pipe.SMembers(ctx, getExternalActionsKey(typeAction)),
			actionGroupsCmd:    pipe.SMembers(ctx, getActionGroupKey(typeAction)),
		}
	}

	employeeGroupsCmd := pipe.SMembers(ctx, getEmployeeActionGroupKey(employeeID))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get data from redis: %w", err)
	}

	employeeGroups, err := parseEmployeeGroups(employeeGroupsCmd)
	if err != nil {
		return nil, err
	}

	policies := make(map[string]*models.AccessData, len(typeActions))
	for typeAction, commands := range commandsMap {
		externalActions, err := parseExternalActions(commands.externalActionsCmd)
		if err != nil {
			return nil, err
		}

		actionGroups, err := parseActionGroups(commands.actionGroupsCmd)
		if err != nil {
			return nil, err
		}

		policies[typeAction] = &models.AccessData{
			ExternalActions: externalActions,
			ActionGroups:    actionGroups,
		}
	}

	return &models.AccessPoliciesResult{
		Policies:       policies,
		EmployeeGroups: employeeGroups,
	}, nil
}

func parseExternalActions(eaCmd *redis.StringSliceCmd) ([]models.ExternalActionWithCategory, error) {
	raw, err := eaCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get externalActions from cache, err: %w", err)
	}

	externalActions := make([]models.ExternalActionWithCategory, len(raw))
	for i, v := range raw {
		if err := jsoniter.Unmarshal([]byte(v), &externalActions[i]); err != nil {
			return nil, fmt.Errorf("can't unmarshal externalActions from cache, err: %w", err)
		}
	}
	return externalActions, nil
}

func parseActionGroups(agCmd *redis.StringSliceCmd) ([]models.ActionGroupWithCategory, error) {
	raw, err := agCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get actionGroups from cache, err: %w", err)
	}

	actionGroups := make([]models.ActionGroupWithCategory, len(raw))
	for i, v := range raw {
		if err := jsoniter.Unmarshal([]byte(v), &actionGroups[i]); err != nil {
			return nil, fmt.Errorf("can't unmarshal actionGroups from cache, err: %w", err)
		}
	}

	return actionGroups, nil
}

func parseEmployeeGroups(egCmd *redis.StringSliceCmd) ([]int64, error) {
	raw, err := egCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get employeeGroups from cache, err: %w", err)
	}

	employeeGroups := make([]int64, len(raw))
	for i, groupStr := range raw {
		group, err := strconv.ParseInt(groupStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal employeeGroups from cache, err: %w", err)
		}
		employeeGroups[i] = group
	}

	return employeeGroups, nil
}

func getExternalActionsKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_external_actions:%s", typeAction)
}

func getActionGroupKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_group_ids:%s", typeAction)
}

func getEmployeeActionGroupKey(employeeID int64) string {
	return fmt.Sprintf("employee_action_group:%d", employeeID)
}
