package models

type GetSprutsResponse struct {
	Spruts []Sprut `json:"spruts"`
}

type Sprut struct {
	SprutName  string   `json:"sprut_name"`
	SprutColor string   `json:"sprut_color"`
	Offices    []Office `json:"offices"`
}

type Office struct {
	OfficeId    int64   `json:"office_id"`
	City        string  `json:"city"`
	OfficeName  string  `json:"office_name"`
	FullAddress string  `json:"full_address"`
	TypePoint   int64   `json:"type_point"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}
