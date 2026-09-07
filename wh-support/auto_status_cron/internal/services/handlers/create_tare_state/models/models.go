package createtarestatemodels

type ExtTicketInfoForCreateTareState struct {
	StateID    string `mapstructure:"state_id"`
	StateDescr string `mapstructure:"state_descr"`
}

type RequestForTareStateCreate struct {
	StateID   string `json:"state_id"`
	StateDesc string `json:"state_desc"`
	LocLang   string `json:"loc_lang"`
}
