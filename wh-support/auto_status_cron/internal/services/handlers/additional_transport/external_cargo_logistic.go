package additionaltransport

import (
	"context"
	"fmt"

	additionaltransportmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/additional_transport/models"
)

const (
	logisticCargoStatusNew       = "CARGO_STATUS_NEW"
	logisticCargoStatusCanceled  = "CARGO_STATUS_CANCELED"
	logisticCargoStatusDelivered = "CARGO_STATUS_DELIVERED"
)

const (
	InternalTicketActionReject   additionaltransportmodels.InternalTicketAction = "reject"
	InternalTicketActionComplete additionaltransportmodels.InternalTicketAction = "perform"
	InternalTicketActionNone     additionaltransportmodels.InternalTicketAction = "none"
)

const (
	errMsgLogisticTicketAlreadyCanceled   = "Запрос на отклонение заявки отменен. Соответствующая заявка уже отменена в системе логистики."
	errMsgLogisticTicketAlreadyPerforming = "Запрос на отклонение заявки отменен. Соответствующая заявка уже взята в работу в системе логистики."
)

type externalCargoLogistic struct {
	logisticApi LogisticCargoAPI
}

func NewExternalCargoLogistic(logisticApi LogisticCargoAPI) *externalCargoLogistic {
	return &externalCargoLogistic{
		logisticApi: logisticApi,
	}
}

func (e *externalCargoLogistic) ResolveInternalTicketsStateByCargos(ctx context.Context, req additionaltransportmodels.RequestForResolveInternalTicketStateByCargo) ([]additionaltransportmodels.ResponseResolveInternalTicketStateByCargo, error) {
	cargosInfo, err := e.logisticApi.GetStatusLogisticTicketsBatched(ctx, additionaltransportmodels.RequestForGetStatusLogisticTickets{
		CargoIds: req.CargoIds,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get cargo status logistic tickets batched: %w", err)
	}

	result := make([]additionaltransportmodels.ResponseResolveInternalTicketStateByCargo, 0, len(cargosInfo.Cargos))
	for _, cargo := range cargosInfo.Cargos {
		switch {
		case cargo.Status == logisticCargoStatusCanceled:
			result = append(result, additionaltransportmodels.ResponseResolveInternalTicketStateByCargo{
				CargoId:      cargo.CargoId,
				Action:       InternalTicketActionReject,
				CancelReason: cargo.CancelReason,
			})
		case cargo.Shipment != nil && (cargo.Shipment.ArrivedToLoad == true || cargo.Status == logisticCargoStatusDelivered):
			result = append(result, additionaltransportmodels.ResponseResolveInternalTicketStateByCargo{
				CargoId: cargo.CargoId,
				Action:  InternalTicketActionComplete,
				PerformInfo: additionaltransportmodels.AddedPerformInfoDriver{
					DriverName:   cargo.Shipment.DriverName,
					LicensePlate: cargo.Shipment.LicensePlate,
				},
			})
		default:
			result = append(result, additionaltransportmodels.ResponseResolveInternalTicketStateByCargo{
				CargoId: cargo.CargoId,
				Action:  InternalTicketActionNone,
			})
		}
	}

	return result, nil
}

func (e *externalCargoLogistic) CanRejectInternalTicket(ctx context.Context, req additionaltransportmodels.RequestForCanRejectInternalTicket) (additionaltransportmodels.ResponseCanRejectInternalTicket, error) {
	cargosInfo, err := e.logisticApi.GetStatusLogisticTicketsBatched(ctx, additionaltransportmodels.RequestForGetStatusLogisticTickets{
		CargoIds: []string{req.CargoId},
	})
	if err != nil {
		return additionaltransportmodels.ResponseCanRejectInternalTicket{}, fmt.Errorf("cannot get cargo status logistic tickets batched: %w", err)
	}

	if len(cargosInfo.Cargos) != 1 {
		return additionaltransportmodels.ResponseCanRejectInternalTicket{}, fmt.Errorf("too many cargos from response")
	}

	switch {
	case cargosInfo.Cargos[0].Status == logisticCargoStatusNew && cargosInfo.Cargos[0].Shipment == nil:
		return additionaltransportmodels.ResponseCanRejectInternalTicket{
			Success: true,
		}, nil
	case cargosInfo.Cargos[0].Status == logisticCargoStatusCanceled:
		return additionaltransportmodels.ResponseCanRejectInternalTicket{
			Success: false,
			Reason:  errMsgLogisticTicketAlreadyCanceled,
		}, nil
	default:
		return additionaltransportmodels.ResponseCanRejectInternalTicket{
			Success: false,
			Reason:  errMsgLogisticTicketAlreadyPerforming,
		}, nil
	}
}

func (e *externalCargoLogistic) CreateLogisticCargo(ctx context.Context, req additionaltransportmodels.RequestForCreateLogisticCargo) (string, error) {
	cargoID, err := e.logisticApi.CreateLogisticTicket(ctx, req)
	if err != nil {
		return "", fmt.Errorf("can't create logistic ticket: %w", err)
	}

	return cargoID, nil
}

func (e *externalCargoLogistic) RejectLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestRejectLogisticTicket) error {
	err := e.logisticApi.RejectLogisticTicket(ctx, req)
	if err != nil {
		return fmt.Errorf("can't reject logistic ticket: %w", err)
	}

	return nil
}

type LogisticCargoAPI interface {
	CreateLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestForCreateLogisticCargo) (string, error)
	RejectLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestRejectLogisticTicket) error
	GetStatusLogisticTicketsBatched(ctx context.Context, req additionaltransportmodels.RequestForGetStatusLogisticTickets) (additionaltransportmodels.ResponseGetStatusLogisticTickets, error)
}
