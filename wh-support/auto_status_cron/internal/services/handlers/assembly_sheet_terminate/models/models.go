package assemblysheetterminatemodels

type ExtTicketInfoForAssemblySheetTerminate struct {
	WhID       WhID       `mapstructure:"wh_id"`
	OfficeID   OfficeID   `mapstructure:"office_id"`
	AsmsheetID int64      `mapstructure:"asmsheet_id"`
	EmployeeID EmployeeID `mapstructure:"employee_id"`
}

type WhID struct {
	ID int64 `mapstructure:"id"`
}

type OfficeID struct {
	ID int64 `mapstructure:"id"`
}

type EmployeeID struct {
	ID int64 `mapstructure:"id"`
}

type RequestDataForAssemblySheetTerminate struct {
	OfficeID         int64 `json:"office_id"`
	WhID             int64 `json:"wh_id"`
	EmployeeID       int64 `json:"employee_id"`
	AsmsheetID       int64 `json:"asmsheet_id"`
	CreateEmployeeID int64 `json:"create_employee_id"`
}
