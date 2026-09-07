package access

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
	mapsutils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

func DetermineAccessibleCategories(employeeGroups []int64, employeeExternalActions []string, accessDataList ...*models.AccessData) []int64 {
	if len(accessDataList) == 0 {
		return nil
	}

	employeeExternalActionsMap := buildEmployeeExternalActionsMap(employeeExternalActions)
	employeeGroupsMap := buildEmployeeGroupsMap(employeeGroups)

	mergedCategoryAccessMap := make(map[int64]struct{})

	for _, accessData := range accessDataList {
		if accessData == nil {
			continue
		}

		// Проверяем доступы через внутренние группы
		addCategoryAccessByGroups(accessData.ActionGroups, employeeGroupsMap, mergedCategoryAccessMap)

		// Проверяем доступы через внешние экшены
		addCategoryAccessByExternalActions(accessData.ExternalActions, employeeExternalActionsMap, mergedCategoryAccessMap)
	}

	return mapsutils.Keys(mergedCategoryAccessMap)
}

func addCategoryAccessByExternalActions(
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

func addCategoryAccessByGroups(
	actionGroups []models.ActionGroupWithCategory,
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

func buildEmployeeGroupsMap(employeeGroups []int64) map[int64]struct{} {
	return buildMap(employeeGroups)
}

func buildEmployeeExternalActionsMap(employeeExternalActions []string) map[string]struct{} {
	return buildMap(employeeExternalActions)
}

func buildMap[K comparable](items []K) map[K]struct{} {
	result := make(map[K]struct{}, len(items))
	for _, item := range items {
		result[item] = struct{}{}
	}
	return result
}
