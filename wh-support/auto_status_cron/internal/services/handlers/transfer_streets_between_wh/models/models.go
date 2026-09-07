package transferstreetsbetweenwhmodels

type ExtTicketInfoForTransferStreetsBetweenWh struct {
	OfficeId struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
	WhIdOld struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id_old"`
	WhIdNew struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id_new"`
	StreetsSections []StreetsSections `mapstructure:"streets_sections"`
}

type StreetsSections struct {
	Stage     Stage     `mapstructure:"stage"`
	Streets   Streets   `mapstructure:"streets"`
	NewPart   NewPart   `mapstructure:"new_part"`
	OldPart   OldPart   `mapstructure:"old_part"`
	Sections  Sections  `mapstructure:"sections"`
	IsReverse IsReverse `mapstructure:"is_reverse"`
}

type Stage struct {
	Value struct {
		Id int64 `mapstructure:"id"`
	} `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type Streets struct {
	Value         []int64 `json:"value" mapstructure:"value"`
	OrderID       int64   `json:"order_id" mapstructure:"order_id"`
	FrontDataName string  `json:"front_data_name" mapstructure:"front_data_name"`
}

type NewPart struct {
	Value struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type OldPart struct {
	Value struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type Sections struct {
	Value         []int64 `json:"value" mapstructure:"value"`
	OrderID       int64   `json:"order_id" mapstructure:"order_id"`
	FrontDataName string  `json:"front_data_name" mapstructure:"front_data_name"`
}

type IsReverse struct {
	Value         bool   `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type RequestForTransferWhForStock struct {
	OfficeID     int64 `json:"office_id"`
	OldWhID      int64 `json:"old_wh_id"`
	NewWhID      int64 `json:"new_wh_id"`
	Stage        int64 `json:"stage"`
	OldPart      int64 `json:"old_part"`
	NewPart      int64 `json:"new_part"`
	StreetStart  int64 `json:"street_start"`
	StreetEnd    int64 `json:"street_end"`
	SectionStart int64 `json:"section_start"`
	SectionEnd   int64 `json:"section_end"`
	IsReverse    bool  `json:"is_reverse"`
	EmployeeID   int64 `json:"employee_id"`
}

type ErrCommentValue struct {
	Stage        int64 `json:"stage"`
	StreetStart  int64 `json:"street_start"`
	StreetEnd    int64 `json:"street_end"`
	SectionStart int64 `json:"section_start"`
	SectionEnd   int64 `json:"section_end"`
}
