package models

type BandAttachmentParams struct {
	TicketID             int64
	Message              string
	SenderEmployeeName   string
	CreatedDt            string
	EmployeeLabels       map[int64]string
	DefaultEmployeeLabel string
}
