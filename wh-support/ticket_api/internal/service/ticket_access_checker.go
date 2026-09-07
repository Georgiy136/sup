package service

import (
	"context"
	"fmt"
	"slices"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/access"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
)

type TicketAccessChecker struct {
	redis    RedisClientInterface
	resource ResourceEmployeeAccessClientInterface
}

func NewTicketAccessChecker(redis RedisClientInterface, resource ResourceEmployeeAccessClientInterface) *TicketAccessChecker {
	return &TicketAccessChecker{
		redis:    redis,
		resource: resource,
	}
}

func (t *TicketAccessChecker) HasEmployeeTicketAccess(ctx context.Context, employeeID, ticketID int64, typeActions ...string) (bool, error) {
	ticketInfo, err := getTicketCommonInfo(ticketID)
	if err != nil {
		return false, fmt.Errorf("can't get ticket info: %w", err)
	}
	if ticketInfo == nil {
		return false, fmt.Errorf("%w: %d", common.ErrTicketNotFound, ticketID)
	}

	if ticketInfo.CreateEmployeeID == employeeID ||
		slices.Contains(ticketInfo.FavouriteEmployeeIDs, employeeID) ||
		slices.Contains(ticketInfo.GroupEmployeeIDs, employeeID) {
		return true, nil
	}

	categories, err := t.GetAccessibleCategories(ctx, employeeID, typeActions...)
	if err != nil {
		return false, fmt.Errorf("can't get accessible categories: %w", err)
	}

	return slices.Contains(categories, ticketInfo.CategoryID), nil
}

func (t *TicketAccessChecker) GetCategoryStatusAccessForTicket(
	ctx context.Context,
	employeeID int64,
	ticketID int64,
	typeActions ...string,
) ([]models.CategoryStatusAccess, error) {
	ticketInfo, err := getTicketCommonInfo(ticketID)
	if err != nil {
		return nil, fmt.Errorf("can't get ticket info: %w", err)
	}
	if ticketInfo == nil {
		return nil, fmt.Errorf("%w: %d", common.ErrTicketNotFound, ticketID)
	}

	categoryStatusAccess, err := t.GetEmployeeCategoryStatusAccess(ctx, employeeID, typeActions...)
	if err != nil {
		return nil, fmt.Errorf("can't get employee category status access: %w", err)
	}

	filtered := filterCategoryStatusAccessByCategory(categoryStatusAccess, ticketInfo.CategoryID)
	if len(filtered) == 0 {
		return nil, fmt.Errorf(
			"%w for ticket: %d, category: %d",
			common.ErrNoCategoryStatusAccess,
			ticketID,
			ticketInfo.CategoryID,
		)
	}

	return filtered, nil
}

func (t *TicketAccessChecker) GetEmployeeCategoryStatusAccess(ctx context.Context, employeeID int64, typeActions ...string) ([]models.CategoryStatusAccess, error) {
	accessPoliciesResult, err := t.redis.GetAccessPolicies(ctx, employeeID, typeActions...)
	if err != nil {
		return nil, fmt.Errorf("can't get access policies from cache: %w", err)
	}

	externalActionsByEmployeeID, err := t.resource.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("error get all external access actions by employee ID: %w", err)
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPoliciesResult.EmployeeGroups) == 0 {
		return nil, nil
	}

	return access.DetermineCategoryStatusAccess(
		accessPoliciesResult.EmployeeGroups,
		externalActionsByEmployeeID,
		accessDataListFromPolicies(accessPoliciesResult.Policies, typeActions...)...,
	), nil
}

func (t *TicketAccessChecker) GetAccessibleCategories(ctx context.Context, employeeID int64, typeActions ...string) ([]int64, error) {
	accessPoliciesResult, err := t.redis.GetAccessPolicies(ctx, employeeID, typeActions...)
	if err != nil {
		return nil, fmt.Errorf("can't get resources from cache: %w", err)
	}

	externalActionsByEmployeeID, err := t.resource.GetAccessActionsByEmployeeID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("error get all access actions by employee ID: %w", err)
	}

	if len(externalActionsByEmployeeID) == 0 && len(accessPoliciesResult.EmployeeGroups) == 0 {
		return nil, nil
	}

	return access.DetermineAccessibleCategories(
		accessPoliciesResult.EmployeeGroups,
		externalActionsByEmployeeID,
		accessDataListFromPolicies(accessPoliciesResult.Policies, typeActions...)...,
	), nil
}

func accessDataListFromPolicies(policies map[string]*models.AccessData, typeActions ...string) []*models.AccessData {
	list := make([]*models.AccessData, 0, len(typeActions))
	for _, typeAction := range typeActions {
		list = append(list, policies[typeAction])
	}
	return list
}

func filterCategoryStatusAccessByCategory(accessList []models.CategoryStatusAccess, categoryID int64) []models.CategoryStatusAccess {
	result := make([]models.CategoryStatusAccess, 0, len(accessList))
	for _, item := range accessList {
		if item.CategoryID == categoryID && item.StatusID != nil {
			result = append(result, item)
		}
	}
	return result
}
