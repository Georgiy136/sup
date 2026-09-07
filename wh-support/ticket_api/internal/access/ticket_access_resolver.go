package access

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	mapsutils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

func DetermineCategoryStatusAccess(employeeGroups []int64, employeeExternalActions []string, accessDataList ...*models.AccessData) []models.CategoryStatusAccess {
	if len(accessDataList) == 0 {
		return nil
	}

	employeeExternalActionsMap := buildEmployeeExternalActionsMap(employeeExternalActions)
	employeeGroupsMap := buildEmployeeGroupsMap(employeeGroups)

	mergedAccessMap := make(map[string]models.CategoryStatusAccess)

	for _, accessData := range accessDataList {
		if accessData == nil {
			continue
		}

		// Проверяем доступы через внутренние группы
		addCategoryStatusAccessByGroups(accessData.ActionGroups, employeeGroupsMap, mergedAccessMap)

		// Проверяем доступы через внешние экшены
		addCategoryStatusAccessByExternalActions(accessData.ExternalActions, employeeExternalActionsMap, mergedAccessMap)
	}

	return mapsutils.Values(mergedAccessMap)
}

func addCategoryStatusAccessByGroups(
	actionGroups []models.ActionGroupWithCategory,
	employeeGroupsMap map[int64]struct{},
	accessMap map[string]models.CategoryStatusAccess,
) {
	if len(actionGroups) == 0 || len(employeeGroupsMap) == 0 {
		return
	}

	for _, actionGroup := range actionGroups {
		if _, ok := employeeGroupsMap[actionGroup.GroupID]; ok {
			key := buildCategoryStatusKey(actionGroup.CategoryID, actionGroup.StatusID)
			accessMap[key] = models.CategoryStatusAccess{
				CategoryID: actionGroup.CategoryID,
				StatusID:   actionGroup.StatusID,
			}
		}
	}
}

func addCategoryStatusAccessByExternalActions(
	externalActions []models.ExternalActionWithCategory,
	employeeExternalActionsMap map[string]struct{},
	accessMap map[string]models.CategoryStatusAccess,
) {
	if len(externalActions) == 0 || len(employeeExternalActionsMap) == 0 {
		return
	}

	for _, externalAction := range externalActions {
		if _, ok := employeeExternalActionsMap[externalAction.ExternalAction]; ok {
			key := buildCategoryStatusKey(externalAction.CategoryID, externalAction.StatusID)
			accessMap[key] = models.CategoryStatusAccess{
				CategoryID: externalAction.CategoryID,
				StatusID:   externalAction.StatusID,
			}
		}
	}
}

func buildCategoryStatusKey(categoryID int64, statusID *string) string {
	if statusID == nil {
		return fmt.Sprintf("%d_null", categoryID)
	}
	return fmt.Sprintf("%d_%s", categoryID, *statusID)
}

func HasAccessByPolicy(
	accessPolicy models.ResourceAccessPolicy,
	employeeGroups []int64,
	employeeExternalActions []string,
) bool {
	if len(employeeGroups) == 0 && len(employeeExternalActions) == 0 {
		return false
	}

	employeeGroupsMap := buildEmployeeGroupsMap(employeeGroups)
	employeeExternalActionsMap := buildEmployeeExternalActionsMap(employeeExternalActions)

	for _, allowedAccessGroup := range accessPolicy.AccessGroups {
		if _, exists := employeeGroupsMap[allowedAccessGroup]; exists {
			return true
		}
	}

	for _, allowedExternalAction := range accessPolicy.ExternalActions {
		if _, exists := employeeExternalActionsMap[allowedExternalAction]; exists {
			return true
		}
	}

	return false
}

func CollectAccessibleCategoriesByGroups(
	categoriesEmployee map[int64]struct{},
	actionGroups []models.ActionGroupWithCategory,
	employeeGroups []int64,
) {
	if len(actionGroups) == 0 || len(employeeGroups) == 0 {
		return
	}

	employeeGroupsMap := buildEmployeeGroupsMap(employeeGroups)

	for _, actionGroup := range actionGroups {
		if _, ok := employeeGroupsMap[actionGroup.GroupID]; ok {
			categoriesEmployee[actionGroup.CategoryID] = struct{}{}
		}
	}
}

func CollectAccessibleCategoriesByExternalActions(
	categoriesEmployee map[int64]struct{},
	externalActions []models.ExternalActionWithCategory,
	employeeExternalActions []string,
) {
	if len(externalActions) == 0 || len(employeeExternalActions) == 0 {
		return
	}

	employeeExternalActionsMap := buildEmployeeExternalActionsMap(employeeExternalActions)

	for _, externalAction := range externalActions {
		if _, ok := employeeExternalActionsMap[externalAction.ExternalAction]; ok {
			categoriesEmployee[externalAction.CategoryID] = struct{}{}
		}
	}
}

func IsAccessPolicyEmpty(resourceAP *models.ResourceAccessPolicy) bool {
	return resourceAP == nil || (len(resourceAP.AccessGroups) == 0 && len(resourceAP.ExternalActions) == 0)
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
