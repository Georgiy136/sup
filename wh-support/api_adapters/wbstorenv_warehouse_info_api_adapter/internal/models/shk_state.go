package models

type GetShkStates struct {
	Data []GetShkState `json:"data"`
}
type GetShkState struct {
	StateID    string `json:"state_id"`
	StateDescr string `json:"state_descr"`
}

type ShkStateStatus struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}
