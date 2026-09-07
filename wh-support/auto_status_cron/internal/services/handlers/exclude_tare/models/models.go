package excludetaremodels

type ExtTicketInfoForExcludeTare struct {
	TareExcel      []TareFromDB `mapstructure:"tare_excel" json:"tare_excel"`
	ExclusionHours int64        `mapstructure:"exclusion_hours" json:"exclusion_hours"`
}
type RequestBodyForExcludeTare struct {
	TareExcel      []Tare `json:"tare_excel"`
	ExclusionHours int64  `json:"exclusion_hours"`
}

type RequestBodyForExcludeTareOnWriteoffApi struct {
	TareExcel      []Tare `json:"tare_excel"`
	ExclusionHours int64  `json:"exclusion_hours"`
	EmployeeID     int64  `json:"employee_id"`
}
type Tare struct {
	TareId   int64  `mapstructure:"tare_id" json:"tare_id"`
	TareType string `mapstructure:"tare_type" json:"tare_type"`
}

type TareFromDB struct {
	TareId   TareId   `mapstructure:"tare_id" json:"tare_id"`
	TareType TareType `mapstructure:"tare_type" json:"tare_type"`
}

type TareId struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type TareType struct {
	Value         string `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}
