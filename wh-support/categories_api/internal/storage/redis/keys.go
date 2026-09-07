package redis

import "fmt"

func getExternalActionsKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_external_actions:%s", typeAction)
}

func getActionGroupKey(typeAction string) string {
	return fmt.Sprintf("access:type_action_group_ids:%s", typeAction)
}

func getEmployeeActionKey(employeeID int64) string {
	return fmt.Sprintf("employee_action_group:%d", employeeID)
}
