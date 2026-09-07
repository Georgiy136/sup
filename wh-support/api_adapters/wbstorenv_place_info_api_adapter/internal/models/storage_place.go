package models

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}

type StoragePlaceTypes struct {
	Data []StoragePlaceType `json:"data"`
}

type StoragePlaceType struct {
	PlaceTypeID   int64  `json:"place_type_id"`
	PlaceTypeName string `json:"place_type_name"`
}
