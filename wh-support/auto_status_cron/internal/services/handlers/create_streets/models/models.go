package createstreetsmodels

type ExtTicketInfoForCreateStreets struct {
	WhId struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
	Stages      []Stage `mapstructure:"stages"`
	StreetStart int64   `mapstructure:"street_start"`
	StreetEnd   int64   `mapstructure:"street_end"`
	SectionFrom int64   `mapstructure:"section_from"`
	SectionEnd  int64   `mapstructure:"section_end"`
}

type RequestForCreateStreet struct {
	OfficeID     int64   `json:"office_id"`
	WhID         int64   `json:"wh_id"`
	Stages       []int64 `json:"stage"`
	StreetStart  int64   `json:"street_start"`
	StreetEnd    int64   `json:"street_end"`
	SectionStart int64   `json:"section_start"`
	SectionEnd   int64   `json:"section_end"`
	EmployeeID   int64   `json:"employee_id"`
}

type Stage struct {
	ID int64 `mapstructure:"id"`
}
