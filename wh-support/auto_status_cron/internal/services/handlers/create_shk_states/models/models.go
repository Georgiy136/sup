package createshkstatesmodels

type ExtTicketInfoForCreateShkStates struct {
	StateID   string `mapstructure:"state_id"`
	StateName string `mapstructure:"state_name"`
	ProcessID struct {
		ID string `mapstructure:"id"`
	} `mapstructure:"process_id"`
}

type RequestForCreateShkStates struct {
	StateID    string `json:"state_id"`
	StateDescr string `json:"state_descr"`
	ProcessID  string `json:"process_id"`
	LocLang    string `json:"loc_lang"`
}
