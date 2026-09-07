package inclusionwhorstagesinsale

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	inclusionwhorstagesinsalemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_wh_or_stages_in_sale/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerInclusionWhOrStagesInSale"

	statusIdInclusionWhInSaleEN1     = "EN1"
	statusIdInclusionStagesInSaleEN2 = "EN2"
)

type TicketHandlerInclusionWhOrStagesInSale struct {
	repo               services.HandlerTicketsRepo
	exclusionStagesApi exclusionStagesApi
}

func NewTicketHandlerInclusionWhOrStagesInSale(repo services.HandlerTicketsRepo, exclusionStagesApi exclusionStagesApi) *TicketHandlerInclusionWhOrStagesInSale {
	return &TicketHandlerInclusionWhOrStagesInSale{
		repo:               repo,
		exclusionStagesApi: exclusionStagesApi,
	}
}
func (h *TicketHandlerInclusionWhOrStagesInSale) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdInclusionWhInSaleEN1:
		if err := h.includeWhInSale(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't include wh in sale: %w", handlerName, err)
		}
	case statusIdInclusionStagesInSaleEN2:
		if err := h.includeStagesInSale(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't include stages in sale: %w", handlerName, err)
		}
	default:
		return fmt.Errorf("[%s] invalid status ID: %s", handlerName, ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerInclusionWhOrStagesInSale) includeWhInSale(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionwhorstagesinsalemodels.ExtTicketInfoForInclusionWh

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	req := commonhandlersmodels.RequestForStageOrWhExclusionFromSaleUpdateV001{
		WhId:       ext.WhID.ID,
		OfficeId:   ext.OfficeID.ID,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsExcluded: false,
	}

	return h.updateStagesInclusionInSale(ctx, ticketInfo.TicketID, req)
}

func (h *TicketHandlerInclusionWhOrStagesInSale) includeStagesInSale(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionwhorstagesinsalemodels.ExtTicketInfoForInclusionStages

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	stages := make([]int64, len(ext.Stages))
	for i := range ext.Stages {
		stages[i] = ext.Stages[i].ID
	}

	req := commonhandlersmodels.RequestForStageOrWhExclusionFromSaleUpdateV001{
		WhId:       ext.WhID.ID,
		OfficeId:   ext.OfficeID.ID,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsExcluded: false,
		Stages:     stages,
	}

	return h.updateStagesInclusionInSale(ctx, ticketInfo.TicketID, req)
}

func (h *TicketHandlerInclusionWhOrStagesInSale) updateStagesInclusionInSale(ctx context.Context, ticketID int64, request commonhandlersmodels.RequestForStageOrWhExclusionFromSaleUpdateV001) error {
	if err := h.exclusionStagesApi.StageOrWhExclusionFromSaleUpdateV001(ctx, request); err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			if rejectErr := h.repo.RejectTicket(ctx, ticketID, errWithMsg.Msg, locMsgs); rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketID, rejectErr)
			}
			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("[%s] can't make request for include in sale: %w", handlerName, err)
	}

	if err := h.repo.PerformTicket(ctx, ticketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketID, rejectErr)
		}
		return fmt.Errorf("[%s] can't perform ticket %d: %w", handlerName, ticketID, err)
	}

	return nil
}

type exclusionStagesApi interface {
	StageOrWhExclusionFromSaleUpdateV001(ctx context.Context, req commonhandlersmodels.RequestForStageOrWhExclusionFromSaleUpdateV001) error
}
