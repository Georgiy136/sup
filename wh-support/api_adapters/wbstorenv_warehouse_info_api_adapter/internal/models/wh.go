package models

type WH struct {
	Data []WHInfo `json:"data"`
}

type WHInfo struct {
	WhID              int64  `json:"wh_id"`
	WhName            string `json:"wh_name"`
	OfficeID          int64  `json:"office_id"`
	IsDefault         bool   `json:"is_default"`
	OffsetHourFromMsk int64  `json:"offset_hour_from_msk"`
	TimeZone          string `json:"time_zone"`
	PlaceNamePrefix   string `json:"place_name_prefix"`
	IsDeleted         bool   `json:"is_deleted"`
}

type WHInfoV2 struct {
	WhID   int64  `json:"wh_id"`
	WhName string `json:"wh_name"`
}

type SelectorWH struct {
	WhID   int64  `json:"id"`
	WhName string `json:"name"`
}

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}
