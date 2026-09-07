package inclusionstreetinsalemodels

type ExtTicketInfoForInclusionStreetInSale struct {
	WhID struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"wh_id"`
	OfficeID struct {
		ID   int64  `mapstructure:"id"`
		Name string `mapstructure:"name"`
	} `mapstructure:"office_id"`
	StreetsWithSections []struct {
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
	} `mapstructure:"streets_with_sections"`
}
