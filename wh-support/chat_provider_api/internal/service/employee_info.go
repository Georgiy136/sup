package service

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/cache/inmemory_backend"
)

type employeeInfo struct {
	client EmployeeNameClient
	cache  *inmemory_backend.CacheBackend
}

type EmployeeNameClient interface {
	GetEmployeeSelfFullName(ctx context.Context, employeeID int64) (string, error)
}

const defaultTTL = 300

func NewEmployeeInfo(client EmployeeNameClient) *employeeInfo {
	return &employeeInfo{
		client: client,
		cache: inmemory_backend.NewCacheBackend(inmemory_backend.Config{
			TTL: defaultTTL,
		}),
	}
}

func (e *employeeInfo) GetEmployeeName(ctx context.Context, employeeID int64) (string, error) {
	if employeeName, exists := e.cache.Get(strconv.FormatInt(employeeID, 10)); exists {
		return string(employeeName), nil
	}

	employeeName, err := e.client.GetEmployeeSelfFullName(ctx, employeeID)
	if err != nil {
		return "", fmt.Errorf("get employee self full name: %w", err)
	}

	e.cache.Set(strconv.FormatInt(employeeID, 10), []byte(employeeName))
	return employeeName, nil
}
