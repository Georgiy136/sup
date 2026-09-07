package inclusionstreetmodels

type ExtTicketInfoForInclusionStreet struct {
	WhId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id"`
	Streets []struct {
		Stage struct {
			ID int64 `mapstructure:"id"`
		} `mapstructure:"stage"`
		Street []int64 `mapstructure:"street"`
	} `mapstructure:"streets"`
	OfficeId struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
}
