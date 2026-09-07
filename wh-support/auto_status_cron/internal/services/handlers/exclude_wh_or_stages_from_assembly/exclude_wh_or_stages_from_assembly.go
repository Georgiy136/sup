package excludewhorstagesfromassembly

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	excludewhorstagesfromassemblymodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_wh_or_stages_from_assembly/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const (
	handlerName = "TicketHandlerExcludeWhOrStagesFromAssembly"

	statusIdExcludeWhFromAssemblyEX1     = "EX1"
	statusIdExcludeStagesFromAssemblyEX2 = "EX2"
)

type TicketHandlerExcludeWhOrStagesFromAssembly struct {
	repo                                 services.HandlerTicketsRepo
	whOrStagesAssemblyExclusionProcessor whOrStagesAssemblyExclusionProcessor
}

func NewTicketHandlerExcludeWhOrStagesFromAssembly(repo services.HandlerTicketsRepo, whOrStagesAssemblyExclusionProcessor whOrStagesAssemblyExclusionProcessor) *TicketHandlerExcludeWhOrStagesFromAssembly {
	return &TicketHandlerExcludeWhOrStagesFromAssembly{
		repo:                                 repo,
		whOrStagesAssemblyExclusionProcessor: whOrStagesAssemblyExclusionProcessor,
	}
}

func (h *TicketHandlerExcludeWhOrStagesFromAssembly) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdExcludeWhFromAssemblyEX1:
		if err := h.excludeWhFromAssembly(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't exclude wh from assembly: %w", handlerName, err)
		}
	case statusIdExcludeStagesFromAssemblyEX2:
		if err := h.excludeStagesFromAssembly(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't exclude stages from assembly: %w", handlerName, err)
		}
	default:
		return fmt.Errorf("[%s] invalid status ID: %s", handlerName, ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerExcludeWhOrStagesFromAssembly) excludeWhFromAssembly(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext excludewhorstagesfromassemblymodels.ExtTicketInfoForExcludeWh

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	return h.whOrStagesAssemblyExclusionProcessor.HandleWhOrStagesExclusionFromAssembly(ctx, commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001{
		WhId:       ext.WhID.ID,
		OfficeId:   ext.OfficeID.ID,
		EmployeeId: ticketInfo.CreateEmployeeID,
		IsExcluded: true,
		TicketID:   ticketInfo.TicketID,
	})
}

func (h *TicketHandlerExcludeWhOrStagesFromAssembly) excludeStagesFromAssembly(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext excludewhorstagesfromassemblymodels.ExtTicketInfoForExcludeStages

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
		IsExcluded: true,
		Stages:     stages,
		TicketID:   ticketInfo.TicketID,
	})
}

type whOrStagesAssemblyExclusionProcessor interface {
	HandleWhOrStagesExclusionFromAssembly(ctx context.Context, req commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001) error
}
