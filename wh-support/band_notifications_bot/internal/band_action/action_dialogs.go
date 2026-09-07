package bandaction

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
)

func (b *BandActionService) openRejectDialog(ctx context.Context, req models.BandActionRequest) error {
	return b.bandBot.OpenInteractiveDialog(ctx, req.TriggerID, models.BandInteractiveDialogConfig{
		CallbackID:       consts.BandRejectDialogCallbackID,
		Title:            consts.BandRejectDialogTitle,
		IntroductionText: fmt.Sprintf(consts.BandRejectDialogIntroFormat, req.TicketID),
		SubmitLabel:      consts.BandRejectSubmitLabel,
		Path:             consts.BandRejectDialogSubmitPath,
		Elements: []models.BandDialogElement{
			{
				DisplayName: consts.BandRejectCommentDisplayName,
				Name:        consts.BandCommentField,
				Type:        models.BandDialogElementTypeTextarea,
				Placeholder: consts.BandRejectCommentPlaceholder,
				Required:    true,
			},
		},
	}, b.buildBandDialogState(req))
}

func (b *BandActionService) openReturnToStatusDialog(ctx context.Context, req models.BandActionRequest) error {
	categoryStatusResp, err := b.ticketRepo.GetCategoryStatusByStatus(ctx, req.CategoryID, req.StatusID)
	if err != nil {
		return fmt.Errorf("can't get category status from db for ticket %d, category_id: %d, status_id: %s, error: %w", req.TicketID, req.CategoryID, req.StatusID, err)
	}
	if len(categoryStatusResp.Data) != 1 {
		return fmt.Errorf("invalid length data from resp db: %d, expected 1", len(categoryStatusResp.Data))
	}

	returnStatuses := categoryStatusResp.Data[0].ReturnStatuses
	if len(returnStatuses) == 0 {
		return fmt.Errorf("empty return statuses for ticket %d: %w", req.TicketID, customerrors.ErrInvalidAction)
	}

	options := make([]*models.BandDialogElementOption, 0, len(returnStatuses))
	for _, status := range returnStatuses {
		options = append(options, &models.BandDialogElementOption{
			Text:  status.ReturnStatusDescription,
			Value: status.ReturnStatusID,
		})
	}

	return b.bandBot.OpenInteractiveDialog(ctx, req.TriggerID, models.BandInteractiveDialogConfig{
		CallbackID:       consts.BandReturnDialogCallbackID,
		Title:            consts.BandReturnDialogTitle,
		IntroductionText: fmt.Sprintf(consts.BandReturnDialogIntroFormat, req.TicketID),
		SubmitLabel:      consts.BandReturnSubmitLabel,
		Path:             consts.BandReturnDialogSubmitPath,
		Elements: []models.BandDialogElement{
			{
				DisplayName: consts.BandReturnStatusDisplayName,
				Name:        consts.BandReturnToStatusIDField,
				Type:        models.BandDialogElementTypeSelect,
				Required:    true,
				Options:     options,
			},
			{
				DisplayName: consts.BandReturnCommentDisplayName,
				Name:        consts.BandCommentField,
				Type:        models.BandDialogElementTypeTextarea,
				Placeholder: consts.BandReturnCommentPlaceholder,
				Required:    true,
			},
		},
	}, b.buildBandDialogState(req))
}

func (b *BandActionService) openAdditionalInfoDialogForAction(ctx context.Context, req models.BandActionRequest, categoryStatusData *ticketmodels.CategoryStatusData) error {
	var actionName, buttonName string
	switch req.Action {
	case consts.ApproveAction:
		actionName = consts.BandApproveActionName
		buttonName = consts.BandApproveButtonName
	case consts.PerformAction:
		actionName = consts.BandPerformActionName
		buttonName = consts.BandPerformButtonName
	}

	elements, err := convertFieldsToDialogElements(categoryStatusData.AddedInformationModel)
	if err != nil {
		return fmt.Errorf("can't convert fields for ticket %d: %w", req.TicketID, err)
	}

	return b.bandBot.OpenInteractiveDialog(ctx, req.TriggerID, models.BandInteractiveDialogConfig{
		CallbackID:       consts.BandAdditionalInfoDialogCallbackID,
		Title:            fmt.Sprintf(consts.BandActionDialogTitleFormat, actionName, req.TicketID),
		IntroductionText: consts.BandAdditionalInfoIntroductionText,
		SubmitLabel:      buttonName,
		Path:             consts.BandAdditionalInfoDialogPath,
		Elements:         elements,
	}, b.buildBandDialogState(req))
}

func (b *BandActionService) buildBandDialogState(req models.BandActionRequest) models.BandDialogState {
	now := time.Now().Unix()
	return models.BandDialogState{
		TicketID:         req.TicketID,
		CategoryID:       req.CategoryID,
		StatusID:         req.StatusID,
		Action:           req.Action,
		Operation:        req.Operation,
		TypeOfEmployeeID: req.TypeOfEmployeeID,
		EmployeeID:       req.EmployeeID,
		ChannelID:        req.ChannelID,
		PostID:           req.PostID,
		Signature: b.actionSigner.Sign(
			strconv.FormatInt(req.TicketID, 10),
			req.Action,
			req.Operation,
			strconv.FormatInt(req.EmployeeID, 10),
			req.TypeOfEmployeeID.String(),
			strconv.FormatInt(now, 10),
		),
		Timestamp: now,
	}
}
