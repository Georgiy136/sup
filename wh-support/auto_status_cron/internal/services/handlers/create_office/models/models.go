package createofficemodels

type ExtTicketInfoForSearchAndAttachSprut struct {
	PhysicalOfficeId int64 `mapstructure:"physical_office_id"`
}

type AddedPerformInfoSprut struct {
	SprutName struct {
		Name string `json:"name"`
	} `json:"sprut_name"`
}

type ExtTicketInfoForCreateOffice struct {
	SprutName string `mapstructure:"sprut_name"`
	OfficeId  struct {
		Id       int64  `mapstructure:"id"`
		Name     string `mapstructure:"name"`
		TypeName string `mapstructure:"type_name"`
	} `mapstructure:"office_id"`
}

type RequestForCreateOffice struct {
	SprutName  string `json:"sprut_name"`
	OfficeId   int64  `json:"office_id"`
	EmployeeId int64  `json:"employee_id"`
}
