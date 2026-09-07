package utils

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/models"

func DedupEmployeeIDsFromEmployeeInfoFromDB(info []models.EmployeeInfoFromDB) []int64 {
	dedupEmployeeIDs := make(map[int64]struct{}, len(info))

	for i := range info {
		dedupEmployeeIDs[info[i].ChEmployeeID] = struct{}{}
		dedupEmployeeIDs[info[i].EmployeeId] = struct{}{}
	}

	out := make([]int64, 0, len(dedupEmployeeIDs))
	for k := range dedupEmployeeIDs {
		out = append(out, k)
	}

	return out
}
