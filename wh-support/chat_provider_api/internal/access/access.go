package access

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"

func HasAccessByActionGroups(
	categoryID int64,
	statusID string,
	employeeGroups []int64,
	actionGroups []models.ActionGroupWithCategory,
) bool {
	if len(employeeGroups) == 0 || len(actionGroups) == 0 {
		return false
	}

	employeeGroupsMap := make(map[int64]struct{}, len(employeeGroups))
	for _, groupID := range employeeGroups {
		employeeGroupsMap[groupID] = struct{}{}
	}

	for _, actionGroup := range actionGroups {
		if actionGroup.CategoryID != categoryID {
			continue
		}
		if actionGroup.StatusID == nil || statusID != *actionGroup.StatusID {
			continue
		}
		if _, ok := employeeGroupsMap[actionGroup.GroupID]; ok {
			return true
		}
	}
	return false
}
