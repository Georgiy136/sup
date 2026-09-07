package service

import (
	"context"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/models"
)

type AccessPolicyCache interface {
	GetAccessData(ctx context.Context, typeAction string, employeeID int64) (*models.AccessData, error)
}

type ResourceEmployeeAccessClient interface {
	GetAccessActionsByEmployeeID(employeeID int64) ([]string, error)
}
