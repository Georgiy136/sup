package createinventtaskmodelsnew

type ExtTicketInfoCreateInventTaskNoWithdrawalBySections struct {
	Stage struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"stage"`
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	Streets  []int64 `mapstructure:"streets"`
	Sections []struct {
		SectionsRange struct {
			Value []int64 `mapstructure:"value"`
		} `mapstructure:"sections_range"`
	} `mapstructure:"sections"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type ExtTicketInfoCreateInventTaskWithWithdrawalBySections struct {
	Stage struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"stage"`
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	Streets  []int64 `mapstructure:"streets"`
	Sections []struct {
		SectionsRange struct {
			Value []int64 `mapstructure:"value"`
		} `mapstructure:"sections_range"`
	} `mapstructure:"sections"`
	Racks    []int64 `mapstructure:"racks"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type RequestDataForCreateInventTask struct {
	OfficeId   int64   `json:"office_id"`
	WhId       int64   `json:"wh_id"`
	Stage      int64   `json:"stage"`
	Streets    []int64 `json:"streets"`
	Sections   []int64 `json:"sections"`
	PlaceIDs   []int64 `json:"place_ids"`
	Racks      []int64 `json:"racks"`
	Priority   int64   `json:"priority"`
	EmployeeId int64   `json:"employee_id"`
}

type ExtTicketInfoCreateInventTaskWithWithdrawalByPlaceIDs struct {
	Stage struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"stage"`
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	Places []struct {
		PlaceId int64 `mapstructure:"id"`
	} `mapstructure:"places"`
	Street   int64 `mapstructure:"street"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}
