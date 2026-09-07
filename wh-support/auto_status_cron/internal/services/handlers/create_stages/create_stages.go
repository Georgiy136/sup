package createstages

import (
	"context"
	"fmt"

	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	createstagesmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_stages/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const (
	statusIdCreateStages = "EXS"
	statusIdCreateParts  = "EXP"
)

type TicketHandlerCreateStages struct {
	stageBuilder stageBuilder
}

func NewTicketHandlerCreateStages(stageBuilder stageBuilder) *TicketHandlerCreateStages {
	return &TicketHandlerCreateStages{
		stageBuilder: stageBuilder,
	}
}

func (s *TicketHandlerCreateStages) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdCreateStages:
		err := s.createStages(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create stages: %w", err)
		}
	case statusIdCreateParts:
		err := s.createParts(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create parts: %v", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (s *TicketHandlerCreateStages) createStages(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstagesmodels.ExtTicketInfoForCreateStages

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	err = s.stageBuilder.CreateStages(ctx, commonhandlersmodels.HandlerRequestForCreateStages{
		OfficeID:   ext.OfficeId,
		WhID:       ext.WhId.Id,
		Stages:     ext.Stages,
		TicketID:   ticketInfo.TicketID,
		EmployeeID: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		return fmt.Errorf("can't create stages: %w", err)
	}

	return nil
}

func (s *TicketHandlerCreateStages) createParts(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstagesmodels.ExtTicketInfoForCreateParts

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	err = s.stageBuilder.CreateParts(ctx, commonhandlersmodels.HandlerRequestForCreateParts{
		OfficeID:   ext.OfficeId,
		WhID:       ext.WhId.Id,
		Stages:     ext.Stages,
		TicketID:   ticketInfo.TicketID,
		EmployeeID: ticketInfo.CreateEmployeeID,
		PartName:   ext.PartName,
	})
	if err != nil {
		return fmt.Errorf("can't create parts: %w", err)
	}

	return nil
}

type stageBuilder interface {
	CreateStages(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateStages) error
	CreateParts(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateParts) error
}
