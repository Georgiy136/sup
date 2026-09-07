package utils

func DeduplicateNumsSlice(in []int64) []int64 {
	existMap := make(map[int64]struct{})

	for _, employee := range in {
		if _, ok := existMap[employee]; !ok {
			existMap[employee] = struct{}{}
		}
	}

	res := make([]int64, 0, len(existMap))

	for employeeID := range existMap {
		res = append(res, employeeID)
	}

	return res
}
