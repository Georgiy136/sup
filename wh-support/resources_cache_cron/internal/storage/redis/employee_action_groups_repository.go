package redis

import (
	"context"
	"fmt"
	"strconv"
)

type EmployeeActionGroupsRepository struct {
	*RedisClient
}

func NewEmployeeActionGroupsRepository(client *RedisClient) *EmployeeActionGroupsRepository {
	return &EmployeeActionGroupsRepository{client}
}

func (r *EmployeeActionGroupsRepository) AddGroupToEmployee(ctx context.Context, employeeID, groupID int64) error {
	key := EmployeeActionGroupKey(employeeID)

	err := r.rCli.SAdd(ctx, key, strconv.FormatInt(groupID, 10)).Err()
	if err != nil {
		return fmt.Errorf("can't add group to employee in redis, err: %w", err)
	}

	return nil
}

func (r *EmployeeActionGroupsRepository) RemoveGroupFromEmployee(ctx context.Context, employeeID, groupID int64) error {
	key := EmployeeActionGroupKey(employeeID)

	err := r.rCli.SRem(ctx, key, strconv.FormatInt(groupID, 10)).Err()
	if err != nil {
		return fmt.Errorf("can't remove group from employee in redis, err: %w", err)
	}

	return nil
}
