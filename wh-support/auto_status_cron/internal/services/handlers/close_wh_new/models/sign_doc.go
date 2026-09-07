package closewhmodels

type ExtTicketInfoForGenerateSignDoc struct {
	OfficeType struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_type"`
	WhId struct {
		Id   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id"`
	TotalPrice float64 `mapstructure:"total_price_sum"`
	TotalCount int64   `mapstructure:"total_count_goods_id"`
}
