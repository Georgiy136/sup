package inclusionstreetinsale

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	inclusionstreetmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_street_in_sale/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	"github.com/sirupsen/logrus"
)

const (
	handlerName = "TicketHandlerInclusionStreetInSale"

	maxStreetsOnStagePerTicket           = 10
	errMsgTooManyStreetsOnStagePerTicket = "Максимальное количество улиц на одном этаже не должно превышать %d\n"

	maxSectionsPerTicket           = 3000
	errMsgTooManySectionsPerTicket = "Суммарное количество секций по всем улицам не должно превышать %d\n"

	errIncorrectOrderSectionInterval = "Ошибка составления заявки: Проверьте правильность введенных секций для улиц\n"
)

var (
	errTooManyStreetsOnStage = errors.New("too many streets on stage")
	errTooManySections       = errors.New("too many sections")
	errWrongSectionsInterval = errors.New("wrong sections interval")
)

type TicketHandlerInclusionStreetInSale struct {
	repo                    services.HandlerTicketsRepo
	updaterExclusionStreets updaterExclusionStreets
}

func NewTicketHandlerInclusionStreetInSale(repo services.HandlerTicketsRepo, updaterExclusionStreets updaterExclusionStreets) *TicketHandlerInclusionStreetInSale {
	return &TicketHandlerInclusionStreetInSale{
		repo:                    repo,
		updaterExclusionStreets: updaterExclusionStreets,
	}
}

func (h *TicketHandlerInclusionStreetInSale) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext inclusionstreetmodels.ExtTicketInfoForInclusionStreetInSale

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		logrus.Debugf("[%s] can't decode map to structure with err %v unset for ext: %v", handlerName, err, ticketInfo.Ext)
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	var totalLen int
	for i := range ext.StreetsWithSections {
		totalLen += len(ext.StreetsWithSections[i].Street.Value)
	}

	err = h.validStreetsWithSections(ext)
	if err != nil {
		var errMsgSB strings.Builder
		locMsgs := localization.LocalizedErrors{}

		if errors.Is(err, errTooManyStreetsOnStage) {
			errMsgSB.WriteString(fmt.Sprintf(errMsgTooManyStreetsOnStagePerTicket, maxStreetsOnStagePerTicket))
			locMsgs.Add(
				support_err_keys.KeyErrorStreetMaxCountExceeded,
				map[string]string{"max_count": strconv.Itoa(maxStreetsOnStagePerTicket)},
			)
		}

		if errors.Is(err, errTooManySections) {
			errMsgSB.WriteString(fmt.Sprintf(errMsgTooManySectionsPerTicket, maxSectionsPerTicket))
			locMsgs.Add(
				support_err_keys.KeyErrorStreetSectionsExceeded,
				map[string]string{"max_sections": strconv.Itoa(maxSectionsPerTicket)},
			)
		}

		if errors.Is(err, errWrongSectionsInterval) {
			errMsgSB.WriteString(errIncorrectOrderSectionInterval)
			locMsgs.Add(
				support_err_keys.KeyErrorStreetSectionsInvalid,
				nil,
			)
		}

		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsgSB.String(), &locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errMsgSB.String())

		return nil
	}

	streets := make([]commonhandlersmodels.StreetsWithSections, 0, totalLen)
	for i := range ext.StreetsWithSections {
		for j := range ext.StreetsWithSections[i].Street.Value {
			streets = append(streets, commonhandlersmodels.StreetsWithSections{
				Stage:       ext.StreetsWithSections[i].Stage.Value.ID,
				Street:      ext.StreetsWithSections[i].Street.Value[j],
				SectionFrom: ext.StreetsWithSections[i].SectionStart.Value,
				SectionTo:   ext.StreetsWithSections[i].SectionEnd.Value,
			})
		}
	}

	err = h.updaterExclusionStreets.StreetExclusionFromSaleUpdateV002(ctx, commonhandlersmodels.HandlerRequestForStreetExclusionFromSaleUpdateV002{
		OfficeId:           ext.OfficeID.ID,
		WhId:               ext.WhID.ID,
		EmployeeId:         ticketInfo.CreateEmployeeID,
		TicketID:           ticketInfo.TicketID,
		Streets:            streets,
		IsExcludedFromSale: false,
		ReplaceOrders:      false,
	})
	if err != nil {
		return fmt.Errorf("[%s] can't exclude street from assembly err: %w", handlerName, err)
	}

	return nil
}

func (h *TicketHandlerInclusionStreetInSale) validStreetsWithSections(ext inclusionstreetmodels.ExtTicketInfoForInclusionStreetInSale) error {
	var (
		errs                  []error
		sumSections           int64
		tooManyStreetsOnStage bool
	)

	for i := range ext.StreetsWithSections {
		if len(ext.StreetsWithSections[i].Street.Value) > maxStreetsOnStagePerTicket && !tooManyStreetsOnStage {
			errs = append(errs, errTooManyStreetsOnStage)
			tooManyStreetsOnStage = true
		}

		if ext.StreetsWithSections[i].SectionEnd.Value < ext.StreetsWithSections[i].SectionStart.Value {
			return errWrongSectionsInterval
		}

		sumSections += int64(len(ext.StreetsWithSections[i].Street.Value)) * (ext.StreetsWithSections[i].SectionEnd.Value - ext.StreetsWithSections[i].SectionStart.Value + 1)
	}

	if sumSections > maxSectionsPerTicket {
		errs = append(errs, errTooManySections)
	}

	return errors.Join(errs...)
}

type updaterExclusionStreets interface {
	StreetExclusionFromSaleUpdateV002(ctx context.Context, req commonhandlersmodels.HandlerRequestForStreetExclusionFromSaleUpdateV002) error
}
