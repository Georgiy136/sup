package excludestreetforsalemodels

type ExtTicketInfoForOperationBySections struct {
	WhID                WhID                  `mapstructure:"wh_id"`
	OfficeID            OfficeID              `mapstructure:"office_id"`
	StreetsWithSections []StreetsWithSections `mapstructure:"streets_with_sections"`
}

type ExtTicketInfoForOperationBySectionsV002 struct {
	WhID                WhID                  `mapstructure:"wh_id"`
	OfficeID            OfficeID              `mapstructure:"office_id"`
	StreetsWithSections []StreetsWithSections `mapstructure:"streets_with_sections"`
	ReplaceOrders       bool                  `mapstructure:"replace_orders"`
}
type WhID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type OfficeID struct {
	ID   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type StreetsWithSections struct {
	Stage struct {
		Value struct {
			ID int64 `mapstructure:"id"`
		} `mapstructure:"value"`
	} `mapstructure:"stage"`
	Street struct {
		Value []int64 `mapstructure:"value"`
	} `mapstructure:"street"`
	SectionEnd struct {
		Value int64 `mapstructure:"value"`
	} `mapstructure:"section_end"`
	SectionStart struct {
		Value int64 `mapstructure:"value"`
	} `mapstructure:"section_start"`
}
