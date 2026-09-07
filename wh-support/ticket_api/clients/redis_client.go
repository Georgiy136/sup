package clients

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
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

func (r *RedisClient) GetAccessPolicy(ctx context.Context, employeeID int64, typeAction string) (*models.AccessPolicyResult, error) {
	pipe := r.Client.Pipeline()

	externalActionsCmd := pipe.SMembers(ctx, getExternalActionsKey(typeAction))
	actionGroupsCmd := pipe.SMembers(ctx, getActionGroupKey(typeAction))
	employeeGroupsCmd := pipe.SMembers(ctx, getEmployeeActionGroupKey(employeeID))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get data from redis: %w", err)
	}

	externalActions, err := parseExternalActions(externalActionsCmd)
	if err != nil {
		return nil, err
	}

	actionGroups, err := parseActionGroups(actionGroupsCmd)
	if err != nil {
		return nil, err
	}

	employeeGroups, err := parseEmployeeGroups(employeeGroupsCmd)
	if err != nil {
		return nil, err
	}

	return &models.AccessPolicyResult{
		AccessData: &models.AccessData{
			ExternalActions: externalActions,
			ActionGroups:    actionGroups,
		},
		EmployeeGroups: employeeGroups,
	}, nil
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

func (r *RedisClient) GetAccessGroupsByEmployee(ctx context.Context, employeeID int64) ([]int64, error) {
	values, err := r.SMembers(ctx, getEmployeeActionGroupKey(employeeID)).Result()
	if err != nil {
		return nil, fmt.Errorf("can't get set from redis: %w", err)
	}

	if len(values) == 0 {
		return nil, nil
	}

	groups := make([]int64, 0, len(values))
	for _, v := range values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't parse group id \"%s\": %w", v, err)
		}
		groups = append(groups, id)
	}

	return groups, nil
}

func (r *RedisClient) GetResourceAccessPolicy(ctx context.Context, resource models.ResourceCompositeKey) (*models.ResourceAccessPolicy, error) {
	pipe := r.Client.Pipeline()

	externalActionsCmd := pipe.SMembers(ctx, getAccessPolicyForExternalActionsKey(resource))
	accessGroupsCmd := pipe.SMembers(ctx, getAccessPolicyForActionGroupKey(resource))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't exec redis pipe: %w", err)
	}

	externalActionsResult, err := externalActionsCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get external actions from redis: %w", err)
	}

	accessGroupsResult, err := accessGroupsCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get access groups from redis: %w", err)
	}

	accessGroups := make([]int64, 0, len(accessGroupsResult))
	for _, groupStr := range accessGroupsResult {
		group, err := strconv.ParseInt(groupStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't parse group id \"%s\": %w", groupStr, err)
		}

		accessGroups = append(accessGroups, group)
	}

	return &models.ResourceAccessPolicy{
		ExternalActions: externalActionsResult,
		AccessGroups:    accessGroups,
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

func getAccessPolicyForActionGroupKey(resource models.ResourceCompositeKey) string {
	return fmt.Sprintf("access:group_ids:%d_%s_%s", resource.CategoryID, resource.TypeAction, formatOmitemptyPartKey(resource.StatusID))
}

func getAccessPolicyForExternalActionsKey(resource models.ResourceCompositeKey) string {
	return fmt.Sprintf("access:external_actions:%d_%s_%s", resource.CategoryID, resource.TypeAction, formatOmitemptyPartKey(resource.StatusID))
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

func formatOmitemptyPartKey[T any](v *T) any {
	if v == nil {
		return "null"
	}
	return *v
}

func (r *RedisClient) GetChatActivities(ctx context.Context, employeeID int64, ticketIDs []int64) (map[int64]*models.ChatActivityData, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}

	pipe := r.Client.Pipeline()

	type cmdPair struct {
		lastActivityCmd     *redis.StringCmd
		lastUserActivityCmd *redis.StringCmd
	}

	cmds := make(map[int64]*cmdPair, len(ticketIDs))
	for _, ticketID := range ticketIDs {
		cmds[ticketID] = &cmdPair{
			lastActivityCmd:     pipe.Get(ctx, getLastChatActivityKey(ticketID)),
			lastUserActivityCmd: pipe.Get(ctx, getLastChatViewByUserKey(employeeID, ticketID)),
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("can't get chat activities from redis: %w", err)
	}

	result := make(map[int64]*models.ChatActivityData, len(ticketIDs))
	for ticketID, pair := range cmds {
		data := &models.ChatActivityData{}

		if val, err := pair.lastActivityCmd.Result(); err == nil {
			data.LastActivityAt = val
		}
		if val, err := pair.lastUserActivityCmd.Result(); err == nil {
			data.LastUserActivityAt = val
		}

		result[ticketID] = data
	}

	return result, nil
}

func getLastChatActivityKey(ticketID int64) string {
	return fmt.Sprintf("ticket_chat:%d:last_activity", ticketID)
}

func getLastChatViewByUserKey(employeeID, ticketID int64) string {
	return fmt.Sprintf("chat_user:%d:ticket_id:%d:last_seen", employeeID, ticketID)
}
