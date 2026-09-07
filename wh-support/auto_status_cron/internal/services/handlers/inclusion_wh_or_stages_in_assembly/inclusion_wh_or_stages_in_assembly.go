package inclusionwhorstagesinassembly

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	inclusionwhorstagesinassemblymodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_wh_or_stages_in_assembly/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const (
	handlerName = "TicketHandlerInclusionWhOrStagesInAssembly"

	statusIdInclusionWhInAssemblyEX1     = "EX1"
	statusIdInclusionStagesInAssemblyEX2 = "EX2"
)

type TicketHandlerInclusionWhOrStagesInAssembly struct {
	repo                                 services.HandlerTicketsRepo
	whOrStagesAssemblyExclusionProcessor whOrStagesAssemblyExclusionProcessor
}

func NewTicketHandlerInclusionWhOrStagesInAssembly(repo services.HandlerTicketsRepo, whOrStagesAssemblyExclusionProcessor whOrStagesAssemblyExclusionProcessor) *TicketHandlerInclusionWhOrStagesInAssembly {
	return &TicketHandlerInclusionWhOrStagesInAssembly{
		repo:                                 repo,
		whOrStagesAssemblyExclusionProcessor: whOrStagesAssemblyExclusionProcessor,
	}
}

func (h *TicketHandlerInclusionWhOrStagesInAssembly) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdInclusionWhInAssemblyEX1:
		if err := h.includeWhInAssembly(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't include wh in assembly: %w", handlerName, err)
		}
	case statusIdInclusionStagesInAssemblyEX2:
		if err := h.includeStagesInAssembly(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't include stages in assembly: %w", handlerName, err)
		}
	default:
		return fmt.Errorf("[%s] invalid status ID: %s", handlerName, ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerInclusionWhOrStagesInAssembly) includeWhInAssembly(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionwhorstagesinassemblymodels.ExtTicketInfoForInclusionWh

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	return h.whOrStagesAssemblyExclusionProcessor.HandleWhOrStagesExclusionFromAssembly(ctx, commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001{
		WhId:       ext.WhID.ID,
		OfficeId:   ext.OfficeID.ID,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsExcluded: false,
		TicketID:   ticketInfo.TicketID,
	})
}

func (h *TicketHandlerInclusionWhOrStagesInAssembly) includeStagesInAssembly(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionwhorstagesinassemblymodels.ExtTicketInfoForInclusionStages

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	stages := make([]int64, len(ext.Stages))
	for i := range ext.Stages {
		stages[i] = ext.Stages[i].ID
	}

	return h.whOrStagesAssemblyExclusionProcessor.HandleWhOrStagesExclusionFromAssembly(ctx, commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001{
		WhId:       ext.WhID.ID,
		OfficeId:   ext.OfficeID.ID,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsExcluded: false,
		Stages:     stages,
		TicketID:   ticketInfo.TicketID,
	})
}

type whOrStagesAssemblyExclusionProcessor interface {
	HandleWhOrStagesExclusionFromAssembly(ctx context.Context, req commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001) error
}
