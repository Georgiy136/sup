package models

type ResponseWithCallBacks struct {
	ApiCallBacks ApiCallbacks `json:"api_callback"`
}

type ApiCallbacks struct {
	Apis    []Apis  `json:"apis"`
	Token   *string `json:"token"`
	Version string  `json:"version"`
}

type Apis struct {
	Key  string   `json:"key"`
	Urls []string `json:"urls"`
}

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}
