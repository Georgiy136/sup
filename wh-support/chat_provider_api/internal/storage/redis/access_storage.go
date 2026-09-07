package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type AccessStorage struct {
	cache *Cache
}

func NewAccessStorage(cache *Cache) *AccessStorage {
	return &AccessStorage{cache: cache}
}

func (a *AccessStorage) GetResourceAccessPolicies(ctx context.Context, categoryID int64, statusID string, typeActions ...string) (map[string][]int64, error) {
	pipe := a.cache.Client.Pipeline()

	commandsMap := make(map[string]*redis.StringSliceCmd, len(typeActions))
	for _, typeAction := range typeActions {
		commandsMap[typeAction] = pipe.SMembers(ctx, getAccessPolicyForActionGroupKey(categoryID, typeAction, statusID))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get data from redis: %w", err)
	}

	policies := make(map[string][]int64, len(typeActions))
	for typeAction, cmd := range commandsMap {
		accessGroups, err := parseAccessGroupIDs(cmd)
		if err != nil {
			return nil, err
		}
		policies[typeAction] = accessGroups
	}

	return policies, nil
}

func (a *AccessStorage) GetEmployeeAccessGroups(ctx context.Context, employeeID int64) ([]int64, error) {
	cmd := a.cache.Client.SMembers(ctx, getEmployeeActionGroupKey(employeeID))

	return parseAccessGroupIDs(cmd)
}

func parseAccessGroupIDs(cmd *redis.StringSliceCmd) ([]int64, error) {
	raw, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("can't get access groups from cache, err: %w", err)
	}

	accessGroups := make([]int64, len(raw))
	for i, groupStr := range raw {
		group, err := strconv.ParseInt(groupStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't parse access group from cache, err: %w", err)
		}
		accessGroups[i] = group
	}

	return accessGroups, nil
}

func getAccessPolicyForActionGroupKey(categoryID int64, typeAction, statusID string) string {
	if statusID == "" {
		statusID = "null"
	}
	return fmt.Sprintf("access:group_ids:%d_%s_%s", categoryID, typeAction, statusID)
}

func getEmployeeActionGroupKey(employeeID int64) string {
	return fmt.Sprintf("employee_action_group:%d", employeeID)
}
