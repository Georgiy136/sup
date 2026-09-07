package createoperationmodels

type ExtTicketInfoCreationOperation struct {
	ProdtypeCode   string `mapstructure:"prodtype_code"`
	ProdtypeName   string `mapstructure:"prodtype_name"`
	ProdtypePartID struct {
		ID int64 `mapstructure:"id"`
	} `mapstructure:"prodtype_part_id"`
	IsUseImproverMultiplier bool     `mapstructure:"is_use_improver_multiplier"`
	RateLimit               *float64 `mapstructure:"rate_limit"`
	IsCredit                bool     `mapstructure:"is_credit"`
}

type RequestDataForProdtypesAddNew struct {
	Data []RequestForProdtypesAddNew `json:"data"`
}

type RequestForProdtypesAddNew struct {
	IsCredit                bool     `json:"is_credit"`
	IsDeleted               bool     `json:"is_deleted"`
	ProdtypeCode            string   `json:"prodtype_code"`
	ProdtypeName            string   `json:"prodtype_name"`
	ProdtypePartID          int64    `json:"prodtype_part_id"`
	IsForTarification       bool     `json:"is_for_tarification"`
	IsUseImproverMultiplier bool     `json:"is_use_improver_multiplier"`
	RateLimit               *float64 `json:"rate_limit,omitempty"`
}
