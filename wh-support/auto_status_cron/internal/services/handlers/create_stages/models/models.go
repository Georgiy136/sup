package createstagesmodels

type ExtTicketInfoForCreateStages struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	OfficeId int64   `mapstructure:"office_id"`
	Stages   []int64 `mapstructure:"stages"`
}

type ExtTicketInfoForCreateParts struct {
	WhId struct {
		Id int64 `mapstructure:"id"`
	} `mapstructure:"wh_id"`
	Stages   []int64 `mapstructure:"stages"`
	OfficeId int64   `mapstructure:"office_id"`
	PartName string  `mapstructure:"part_name"`
}
