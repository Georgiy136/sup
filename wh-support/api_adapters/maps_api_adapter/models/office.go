package models

type GetOfficesResponse struct {
	Offices []struct {
		OfficeId       int64   `json:"office_id"`
		City           string  `json:"city"`
		OfficeName     string  `json:"office_name"`
		FullAddress    string  `json:"full_address"`
		TypePoint      int64   `json:"type_point"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		ParentOfficeId int64   `json:"parent_office_id"`
		LmRouteId      int64   `json:"lm_route_id"`
	} `json:"offices"`
	TotalCount int64 `json:"total_count"`
}
