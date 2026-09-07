package redis

import (
	"context"
	"errors"
	"fmt"

	internalactions "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/internal_actions"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
)

type InternalActionsRepository struct {
	*RedisClient
}

func NewInternalActionsRepository(client *RedisClient) *InternalActionsRepository {
	return &InternalActionsRepository{client}
}

func (c *InternalActionsRepository) SetInternalActionsByEmployeeIDs(ctx context.Context, data ...internalactions.EmployeeResources) error {
	for i := range data {
		bytes, err := jsoniter.Marshal(data[i])
		if err != nil {
			return fmt.Errorf("can't marshal data (employeeId = %v), err: %w", data[i].EmployeeId, err)
		}
		err = c.rCli.Set(ctx, InternalActionsKey(data[i].EmployeeId), bytes, 0).Err()
		if err != nil {
			return fmt.Errorf("can't save data to redis, err: %w", err)
		}
	}
	return nil
}

func (c *InternalActionsRepository) GetInternalActionsByEmployeeID(ctx context.Context, employeeID int64) (*internalactions.EmployeeResources, error) {
	resString, err := c.rCli.Get(ctx, InternalActionsKey(employeeID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get data from redis (employeeID = %v), err: %w", employeeID, err)
	}
	result := new(internalactions.EmployeeResources)
	err = jsoniter.Unmarshal([]byte(resString), result)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data from cache (employeeID = %v), err: %w", employeeID, err)
	}
	return result, nil
}

func (c *InternalActionsRepository) DelInternalActionByEmployeeID(ctx context.Context, employeeID int64) error {
	err := c.rCli.Del(ctx, InternalActionsKey(employeeID)).Err()
	if err != nil {
		return fmt.Errorf("can't delete data from redis (employeeID = %v), err: %w", employeeID, err)
	}
	return nil
}
