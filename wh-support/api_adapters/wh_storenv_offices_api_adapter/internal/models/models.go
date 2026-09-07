package models

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}

type ResponseGetAllTypePoint struct {
	Data []TypePoint `json:"data"`
}

type TypePoint struct {
	TypePoint     int64  `json:"type_point"`
	TypePointName string `json:"type_point_name"`
}
