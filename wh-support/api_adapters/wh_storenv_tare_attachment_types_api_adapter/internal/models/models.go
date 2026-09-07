package models

type ResponseWithMsgError struct {
	IsValid bool   `json:"is_valid"`
	ErrMsg  string `json:"err_msg"`
}
