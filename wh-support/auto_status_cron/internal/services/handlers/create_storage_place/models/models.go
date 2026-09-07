package createstorageplacemodels

type ExtTicketInfoForCreateStoragePlace struct {
	Part      Part      `mapstructure:"part"`
	Racks     []int64   `mapstructure:"racks"`
	Stage     Stage     `mapstructure:"stage"`
	WhID      WhID      `mapstructure:"wh_id"`
	Fields    []int64   `mapstructure:"fields"`
	Street    int64     `mapstructure:"street"`
	Sections  []int64   `mapstructure:"sections"`
	OfficeID  OfficeID  `mapstructure:"office_id"`
	PlaceType PlaceType `mapstructure:"place_type"`
}

type Part struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type Stage struct {
	ID int64 `mapstructure:"id"`
}

type WhID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type OfficeID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type PlaceType struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type RequestDataForCreateStoragePlace struct {
	OfficeID     int64 `json:"office_id"`
	WhID         int64 `json:"wh_id"`
	PlaceTypeID  int64 `json:"place_type_id"`
	Part         int64 `json:"part"`
	Stage        int64 `json:"stage"`
	Street       int64 `json:"street"`
	SectionFirst int64 `json:"section_first"`
	SectionLast  int64 `json:"section_last"`
	RackFirst    int64 `json:"rack_first"`
	RackLast     int64 `json:"rack_last"`
	FieldFirst   int64 `json:"field_first"`
	FieldLast    int64 `json:"field_last"`
	EmployeeID   int64 `json:"employee_id"`
}

type ResponseDataFromCreateStoragePlaceApi struct {
	Data []StoragePlace `json:"data"`
}

type StoragePlace struct {
	ChDt         string `json:"ch_dt"`
	OfficeID     int64  `json:"office_id"`
	WhID         int64  `json:"wh_id"`
	Stage        int64  `json:"stage"`
	Part         int64  `json:"part"`
	Street       int64  `json:"street"`
	Section      int64  `json:"section"`
	Rack         int64  `json:"rack"`
	Field        int64  `json:"field"`
	PlaceID      int64  `json:"place_id"`
	PlaceTypeID  int64  `json:"place_type_id"`
	PlaceName    string `json:"place_name"`
	ChEmployeeID int64  `json:"ch_employee_id"`
}

type AddedPerformInfoPlaces struct {
	Places []Place `json:"places"`
}

type Place struct {
	PlaceID   PlaceID   `json:"place_id"`
	PlaceName PlaceName `json:"place_name"`
}

type PlaceID struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type PlaceName struct {
	Value         string `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}
