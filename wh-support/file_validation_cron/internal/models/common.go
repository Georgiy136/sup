package models

type DataWrapper[T any] struct {
	Data T `json:"data"`
}
