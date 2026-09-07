package models

type StreetsResponse struct {
	Data []StreetsInfo `json:"data"`
}

type StreetsInfo struct {
	WhId         int64 `json:"wh_id"`
	Stage        int64 `json:"stage"`
	Street       int64 `json:"street"`
	SectionStart int64 `json:"section_start"`
	SectionEnd   int64 `json:"section_end"`
}

type StreetsSelector struct {
	Street int64 `json:"id"`
}
