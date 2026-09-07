package closewhmodels

type ExtTicketInfoForCloseWh struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type RequestForCloseWh struct {
	OfficeId   int64 `json:"office_id"`
	WhId       int64 `json:"wh_id"`
	EmployeeId int64 `json:"employee_id"`
}

type ExtTicketInfoForCheckUndeletedStoragePlaces struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type RequestForGetUndeletedStoragePlaces struct {
	OfficeId int64 `json:"office_id"`
	WhId     int64 `json:"wh_id"`
}
