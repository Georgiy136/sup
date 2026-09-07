package inclusionstreet

import (
	"context"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	inclusionstreet "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_street/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"

	"github.com/sirupsen/logrus"
)

const handlerName = "TicketHandlerInclusionStreet"

type TicketHandlerInclusionStreet struct {
	repo                    services.HandlerTicketsRepo
	updaterExclusionStreets updaterExclusionStreets
}

func NewTicketHandlerInclusionStreet(repo services.HandlerTicketsRepo, updaterExclusionStreets updaterExclusionStreets) *TicketHandlerInclusionStreet {
	return &TicketHandlerInclusionStreet{
		repo:                    repo,
		updaterExclusionStreets: updaterExclusionStreets,
	}
}

func (h *TicketHandlerInclusionStreet) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionstreet.ExtTicketInfoForInclusionStreet

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		logrus.Debugf("[%s] can't decode map to structure with err %v unset for ext: %v", handlerName, err, ticketInfo.Ext)
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	var totalLen int
	for i := range ext.Streets {
		totalLen += len(ext.Streets[i].Street)
	}

	streets := make([]commonhandlersmodels.Streets, 0, totalLen)

	for i := range ext.Streets {
		for j := range ext.Streets[i].Street {
			streets = append(streets, commonhandlersmodels.Streets{
				Stage:  ext.Streets[i].Stage.ID,
				Street: ext.Streets[i].Street[j],
			})
		}
	}

	err = h.updaterExclusionStreets.ExclusionUpdateStreet(ctx, commonhandlersmodels.HandlerRequestForExclusionUpdateStreet{
		OfficeId:               ext.OfficeId.ID,
		WhId:                   ext.WhId.ID,
		EmployeeId:             ticketInfo.CreateEmployeeID,
		TicketID:               ticketInfo.TicketID,
		Streets:                streets,
		IsExcludedFromAssembly: false,
	})
	if err != nil {
		return fmt.Errorf("[%s] can't exclude street from assembly err: %w", handlerName, err)
	}

	return nil
}

type updaterExclusionStreets interface {
	ExclusionUpdateStreet(ctx context.Context, body commonhandlersmodels.HandlerRequestForExclusionUpdateStreet) error
}
