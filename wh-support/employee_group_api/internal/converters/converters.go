package converters

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/models"

func ConvertEmployeeInfoInMap(employees []models.EmployeeInfo) map[int64]string {
	res := make(map[int64]string, len(employees))

	for _, employee := range employees {
		res[employee.ID] = employee.Name
	}

	return res
}
