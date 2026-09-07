package access

import (
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/statistics_api/internal/models"
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
