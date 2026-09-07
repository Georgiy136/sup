package handlers

import (
	"context"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
)

type BandActionHandler struct {
	bandActionService bandActionService
	actionSigner      actionSigner
}

func NewBandActionHandler(bandActionService bandActionService, actionSigner actionSigner) *BandActionHandler {
	return &BandActionHandler{bandActionService: bandActionService, actionSigner: actionSigner}
}

type bandActionService interface {
	HandleBandAction(ctx context.Context, req models.BandActionRequest) error
	HandleBandRejectDialog(ctx context.Context, req models.BandActionRequest) error
	HandleBandAdditionalInfoDialog(ctx context.Context, req models.BandActionRequest, filledFields map[string]any, state models.BandDialogState) error
	HandleBandReturnToStatusDialog(ctx context.Context, req models.BandActionRequest) error
}

type actionSigner interface {
	Sign(parts ...string) string
	Compare(signature string, parts ...string) bool
}

func (h *BandActionHandler) isAuthorized(signData models.BandActionContext) bool {
	return h.actionSigner.Compare(signData.Signature,
		strconv.FormatInt(signData.TicketID, 10),
		signData.Action,
		signData.Operation,
		strconv.FormatInt(signData.EmployeeID, 10),
		signData.TypeOfEmployeeID.String(),
		strconv.FormatInt(signData.Timestamp, 10),
	)
}
