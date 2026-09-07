package additionaltransportmodels

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
)

type ExtTicketInfosForCreateLogisticTicket struct {
	Date                string              `mapstructure:"date"`
	Cargo               []CargoSupport      `mapstructure:"cargo"`
	ResponsibleEmployee ResponsibleEmployee `mapstructure:"responsible_id"`
	CargoDesc           string              `mapstructure:"cargo_desc"`
	Department          Department          `mapstructure:"department"`
	OfficeIdDst         Office              `mapstructure:"office_id_dst"`
	OfficeIdSrc         Office              `mapstructure:"office_id_src"`
}

type ExtTicketInfosForWaitLogisticTicket struct {
	CargoId string `mapstructure:"cargo_id"`
}

type ResponsibleEmployee struct {
	Id   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type Department struct {
	Id   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type CargoSupport struct {
	Id   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type Office struct {
	Id   int64  `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type RequestForCreateLogisticCargo struct {
	Cargo CargoLogistic `json:"cargo"`
}

type CargoLogistic struct {
	TicketId              int64  `json:"ticket_id"`
	SrcOfficeId           int64  `json:"src_office_id"`
	DstOfficeId           int64  `json:"dst_office_id"`
	LoadDate              string `json:"load_date"`
	Description           string `json:"description"`
	Content               string `json:"content"`
	ResponsibleEmployeeId string `json:"responsible_employee_id"`
	DepartmentName        string `json:"department_name"`
}

type AddedPerformInfoCargoId struct {
	CargoId string `json:"cargo_id"`
}

type RequestForGetStatusLogisticTickets struct {
	CargoIds []string `json:"cargo_ids"`
}

type ResponseGetStatusLogisticTickets struct {
	Cargos []struct {
		CargoId      string `json:"cargo_id"`
		Status       string `json:"status"`
		CancelReason string `json:"cancel_reason"`
		Shipment     *struct {
			DriverName    string `json:"driver_name"`
			LicensePlate  string `json:"license_plate"`
			ArrivedToLoad bool   `json:"arrived_to_load"`
		} `json:"shipment"`
	} `json:"cargos"`
}

type AddedPerformInfoDriver struct {
	DriverName   string `json:"driver_name"`
	LicensePlate string `json:"license_plate"`
}

type RequestForCanRejectInternalTicket struct {
	CargoId string
}

type ResponseCanRejectInternalTicket struct {
	Success   bool
	Reason    string
	LocReason *localization.LocalizedErrors
}

type RequestRejectLogisticTicket struct {
	CargoId string `json:"cargo_id"`
}

type RequestForResolveInternalTicketStateByCargo struct {
	CargoIds []string
}

type InternalTicketAction string

type ResponseResolveInternalTicketStateByCargo struct {
	CargoId      string
	Action       InternalTicketAction
	CancelReason string

	PerformInfo AddedPerformInfoDriver
}
