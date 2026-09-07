package removeremainswrrmodels

type ExtTicketInfoForRemoveRemainsWRR struct {
	ShkList []ShkList `mapstructure:"shk_list"`
}

type ShkList struct {
	NmID   NmID   `json:"nm_id" mapstructure:"nm_id"`
	ShkID  ShkID  `json:"shk_id" mapstructure:"shk_id"`
	ChrtID ChrtID `json:"chrt_id" mapstructure:"chrt_id"`
}

type RequestDataForShksRelease struct {
	RequestShkList []RequestShkList `json:"data"`
}

type RequestShkList struct {
	NmID   int64 `json:"nm_id" mapstructure:"nm_id"`
	ShkID  int64 `json:"shk_id" mapstructure:"shk_id"`
	ChrtID int64 `json:"chrt_id" mapstructure:"chrt_id"`
}

type NmID struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}

type ShkID struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}
type ChrtID struct {
	Value         int64  `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}
