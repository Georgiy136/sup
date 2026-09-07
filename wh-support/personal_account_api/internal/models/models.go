package models

const DatabaseKey = "support_pgx"

type EmployeeNameInfo struct {
	EmployeeID       int64  `json:"employee_id"`
	EmployeeFullName string `json:"employee_name"`
}

type PersonalInfoResp struct {
	Surname    string  `json:"surname"`
	Name       string  `json:"name"`
	Patronymic string  `json:"patronymic"`
	EmployeeId int64   `json:"employee_id"`
	OfficeID   *int64  `json:"office_id"`
	OfficeName *string `json:"office_name"`
	WhID       *int64  `json:"wh_id"`
	WhName     *string `json:"wh_name"`
	Photo      []byte  `json:"photo"`
}

type PersonalInfoRespV2 struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	EmployeeId int64  `json:"employee_id"`
	Photo      []byte `json:"photo"`
}
