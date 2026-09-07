package models

type IDValue interface {
	~int64 | ~string
}

type NamedID[T IDValue] struct {
	ID   T      `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type ID struct {
	ID int64 `mapstructure:"id"`
}

type ValueField[T IDValue] struct {
	Value         T      `json:"value" mapstructure:"value"`
	OrderID       int64  `json:"order_id" mapstructure:"order_id"`
	FrontDataName string `json:"front_data_name" mapstructure:"front_data_name"`
}
