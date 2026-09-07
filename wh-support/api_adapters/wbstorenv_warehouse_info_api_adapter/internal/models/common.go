package models

type ParamsKeys struct {
	AllowedCountryCodes []string `json:"allowed_country_codes" validate:"required,unique,min=1,dive,alphaunicode,len=2"`
}
