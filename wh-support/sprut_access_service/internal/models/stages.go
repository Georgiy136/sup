package models

import "time"

type StagesResponse struct {
	Data StagesInfo `json:"data"`
}

type StagesInfo struct {
	OfficeId int64    `json:"office_id"`
	WhId     int64    `json:"wh_id"`
	Stages   []Stages `json:"stages"`
}

type Stages struct {
	Stage        int64     `json:"stage"`
	ChEmployeeId int64     `json:"ch_employee_id"`
	ChDt         time.Time `json:"ch_dt"`
	Parts        []Parts   `json:"parts"`
}

type Parts struct {
	Part         int64     `json:"part"`
	ChDt         time.Time `json:"ch_dt"`
	ChEmployeeId int64     `json:"ch_employee_id"`
	PartName     string    `json:"part_name"`
}

type StagesSelector struct {
	Stage int64 `json:"id"`
}

type PartsSelector struct {
	Part     int64  `json:"part"`
	PartName string `json:"part_name"`
}
