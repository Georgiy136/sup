package access

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/models"
	mapsutils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

type CategoryAccessResolver struct{}

func NewCategoryAccessResolver() *CategoryAccessResolver {
	return &CategoryAccessResolver{}
}

func (r *CategoryAccessResolver) DetermineAccessibleCategories(
	externalActionsByTypeAction []models.ExternalActionWithCategory,
	actionGroupsByTypeAction []models.ActionGroupInfoWithCategory,
	employeeExternalActions []string,
	employeeGroups []int64,
) []int64 {
	var (
		categoriesMap      = make(map[int64]struct{})
		employeeActionsMap = make(map[string]struct{}, len(employeeExternalActions))
		employeeGroupsMap  = make(map[int64]struct{}, len(employeeGroups))
	)

	for _, action := range employeeExternalActions {
		employeeActionsMap[action] = struct{}{}
	}
	for _, groupID := range employeeGroups {
		employeeGroupsMap[groupID] = struct{}{}
	}

	// Проверяем доступ через внешние экшены
	r.checkExternalActionsAccess(externalActionsByTypeAction, employeeActionsMap, categoriesMap)

	// Проверяем доступ через группы
	r.checkGroupsAccess(actionGroupsByTypeAction, employeeGroupsMap, categoriesMap)

	return mapsutils.Keys(categoriesMap)
}

// checkExternalActionsAccess проверяет доступ к категориям через внешние экшены
func (r *CategoryAccessResolver) checkExternalActionsAccess(
	externalActions []models.ExternalActionWithCategory,
	employeeActionsMap map[string]struct{},
	categoriesMap map[int64]struct{},
) {
	if len(externalActions) == 0 || len(employeeActionsMap) == 0 {
		return
	}

	for _, externalAction := range externalActions {
		if _, exists := employeeActionsMap[externalAction.ExternalAction]; exists {
			categoriesMap[externalAction.CategoryID] = struct{}{}
		}
	}
}

// checkGroupsAccess проверяет доступ к категориям через группы доступа
func (r *CategoryAccessResolver) checkGroupsAccess(
	actionGroups []models.ActionGroupInfoWithCategory,
	employeeGroupsMap map[int64]struct{},
	categoriesMap map[int64]struct{},
) {
	if len(actionGroups) == 0 || len(employeeGroupsMap) == 0 {
		return
	}

	for _, actionGroup := range actionGroups {
		if _, exists := employeeGroupsMap[actionGroup.GroupID]; exists {
			categoriesMap[actionGroup.CategoryID] = struct{}{}
		}
	}
}
