package createwhmodels

type ExtTicketInfoAddBuilding struct {
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	BuildingID string `mapstructure:"building_id_new"`
}

type RequestDataForAddNewBuilding struct {
	IsDel      bool   `json:"is_del"`
	OfficeID   int64  `json:"office_id"`
	BuildingID string `json:"building_id"`
	EmployeeID int64  `json:"employee_id"`
}

type ExtTicketInfoAddPhysicalWhWithCreateBuilding struct {
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	WhName   string `mapstructure:"wh_name"`
	TimeZone struct {
		Id string `mapstructure:"id"`
	} `mapstructure:"time_zone"`
	PlaceNamePrefix          string `mapstructure:"place_name_prefix"`
	BuildingId               string `mapstructure:"building_id_new"`
	IsAssemblyByDeliveryDate bool   `mapstructure:"assembly_by_delivery_date"`
}

type ExtTicketInfoAddPhysicalWhWithSelectBuilding struct {
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	WhName   string `mapstructure:"wh_name"`
	TimeZone struct {
		Id string `mapstructure:"id"`
	} `mapstructure:"time_zone"`
	PlaceNamePrefix string `mapstructure:"place_name_prefix"`
	BuildingId      struct {
		Id string `mapstructure:"id"`
	} `mapstructure:"building_id"`
	IsAssemblyByDeliveryDate bool `mapstructure:"assembly_by_delivery_date"`
}

type ExtTicketInfoAddVirtualWhAW3 struct {
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	WhName   string `mapstructure:"wh_name"`
	TimeZone struct {
		Id string `mapstructure:"id"`
	} `mapstructure:"time_zone"`
	PlaceNamePrefix string `mapstructure:"place_name_prefix"`
	PhysicalWhId    struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"physical_wh_id"`
	IsWhForTmc bool `mapstructure:"tmc_block"`
}

type ExtTicketInfoAddVirtualWhAW4 struct {
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	WhName   string `mapstructure:"wh_name"`
	TimeZone struct {
		Id string `mapstructure:"id"`
	} `mapstructure:"time_zone"`
	PlaceNamePrefix string `mapstructure:"place_name_prefix"`
	PhysicalWhId    struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"physical_wh_id"`
}

type ExtTicketInfoCreateStages struct {
	WhId     int64   `mapstructure:"wh_id"`
	Stages   []int64 `mapstructure:"stages"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
}

type ExtTicketInfoCreateParts struct {
	WhId     int64   `mapstructure:"wh_id"`
	Stages   []int64 `mapstructure:"stages"`
	OfficeId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"office_id"`
	PartName string `mapstructure:"part_name"`
}

type RequestDataForAddNewWh struct {
	WhName                   string  `json:"wh_name"`
	OfficeId                 int64   `json:"office_id"`
	TimeZone                 string  `json:"time_zone"`
	BuildingId               *string `json:"building_id"`
	EmployeeId               int64   `json:"employee_id"`
	PhysicalWhId             *int64  `json:"physical_wh_id"`
	PlaceNamePrefix          string  `json:"place_name_prefix"`
	IsWhForTmc               bool    `json:"is_wh_for_tmc"`
	IsAssemblyByDeliveryDate bool    `json:"assembly_by_delivery_date"`
}

type AddedPerformInfoWh struct {
	WhID int64 `json:"wh_id"`
}
