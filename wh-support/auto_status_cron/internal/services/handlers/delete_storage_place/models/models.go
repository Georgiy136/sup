package deletestorageplacemodels

type ExtTicketInfoForDeleteStoragePlace struct {
	WhID     WhID     `mapstructure:"wh_id"`
	Places   []Place  `mapstructure:"places"`
	OfficeID OfficeID `mapstructure:"office_id"`
}

type WhID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type Place struct {
	PlaceID PlaceID `json:"place_id" mapstructure:"place_id"`
}

type PlaceID struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type OfficeID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type RequestDataForDeleteStoragePlace struct {
	OfficeID   int64   `json:"office_id"`
	PlaceIDs   []int64 `json:"place_ids"`
	EmployeeID int64   `json:"employee_id"`
}
