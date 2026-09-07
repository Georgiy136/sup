package redis

import (
	"fmt"
	"strconv"

	accesspolicy "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/access_policy"
)

func InternalActionsKey(employeeID int64) string {
	return strconv.FormatInt(employeeID, 16)
}

func EmployeeActionGroupKey(employeeID int64) string {
	return fmt.Sprintf("employee_action_group:%d", employeeID)
}

func AccessPolicyForActionGroupKey(resource accesspolicy.ResourceCompositeKey) string {
	return fmt.Sprintf("access:group_ids:%d_%s_%s", resource.CategoryID, resource.TypeAction, formatOmitemptyPartKey(resource.StatusID))
}

func AccessPolicyForExternalActionsKey(resource accesspolicy.ResourceCompositeKey) string {
	return fmt.Sprintf("access:external_actions:%d_%s_%s", resource.CategoryID, resource.TypeAction, formatOmitemptyPartKey(resource.StatusID))
}

func ExternalActionsToTypeActionKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_external_actions:%s", typeAction)
}

func GroupIdsToTypeActionKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_group_ids:%s", typeAction)
}

func formatOmitemptyPartKey[T any](v *T) any {
	if v == nil {
		return "null"
	}
	return *v
}
