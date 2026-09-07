package excludestreetfromsale

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	createinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_invent_task/models"
	excludestreetforsalemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_street_from_sale/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	statusIdExcludeStreetFromSaleEX1                 = "EX1"
	statusIdExcludeStreetFromSaleEX2                 = "EX2"
	statusIdExcludeStreetFromSaleEX3                 = "EX3"
	statusIDCreateInventTaskWithWithdrawalBySections = "EX4"
	statusIDCreateInventTaskNoWithdrawalBySections   = "EX5"

	highPriorityInventTask = 3

	maxStreetsOnStagePerTicket           = 10
	errMsgTooManyStreetsOnStagePerTicket = "Максимальное количество улиц на одном этаже не должно превышать %d\n"

	maxSectionsPerTicket           = 3000
	errMsgTooManySectionsPerTicket = "Суммарное количество секций по всем улицам не должно превышать %d\n"

	defaultErrorComment               = "Внутренняя ошибка"
	startErrorCommentCreateInventTask = "Ошибка при создании инвентаризационного задания:"
	errIncorrectOrderSectionInterval  = "Ошибка составления заявки: Проверьте правильность введенных секций для улиц\n"
)

var (
	errTooManyStreetsOnStage = errors.New("too many streets on stage")
	errTooManySections       = errors.New("too many sections")
	errWrongSectionsInterval = errors.New("wrong sections interval")
)

type TicketHandlerExcludeStreetFromSale struct {
	repo                    services.HandlerTicketsRepo
	updaterExclusionStreets updaterExclusionStreets
	inventTaskCreator       inventTaskCreator
}

func NewTicketHandlerExcludeStreetFromSale(repo services.HandlerTicketsRepo, updaterExclusionStreets updaterExclusionStreets, inventTaskCreator inventTaskCreator) *TicketHandlerExcludeStreetFromSale {
	return &TicketHandlerExcludeStreetFromSale{
		repo:                    repo,
		updaterExclusionStreets: updaterExclusionStreets,
		inventTaskCreator:       inventTaskCreator,
	}
}

func (h *TicketHandlerExcludeStreetFromSale) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIdExcludeStreetFromSaleEX1, statusIdExcludeStreetFromSaleEX2, statusIdExcludeStreetFromSaleEX3:
		if err := h.excludeStreetFromSale(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't exclude street from sale: %w", err)
		}
	case statusIDCreateInventTaskWithWithdrawalBySections:
		if err := h.createInventTaskWithWithdrawalBySections(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't create invent task with withdrawal by sections: %w", err)
		}
	case statusIDCreateInventTaskNoWithdrawalBySections:
		if err := h.createInventTaskNoWithdrawalBySections(ctx, ticketInfo); err != nil {
			return fmt.Errorf("can't create invent task no withdrawal by sections: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerExcludeStreetFromSale) excludeStreetFromSale(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext excludestreetforsalemodels.ExtTicketInfoForOperationBySectionsV002

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		logrus.Debugf("can't decode map to structure with err %v unset for ext: %v", err, ticketInfo.Ext)
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	var totalLen int
	for i := range ext.StreetsWithSections {
		totalLen += len(ext.StreetsWithSections[i].Street.Value)
	}

	err = h.validStreetsWithSections(ext.StreetsWithSections)
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
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errMsgSB.String())
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
		IsExcludedFromSale: true,
		ReplaceOrders:      ext.ReplaceOrders,
	})
	if err != nil {
		return fmt.Errorf("can't exclude street from sale err: %w", err)
	}
	return nil
}

func (h *TicketHandlerExcludeStreetFromSale) validStreetsWithSections(streetsWithSections []excludestreetforsalemodels.StreetsWithSections) error {
	var (
		errs                  []error
		sumSections           int64
		tooManyStreetsOnStage bool
	)

	for i := range streetsWithSections {
		if len(streetsWithSections[i].Street.Value) > maxStreetsOnStagePerTicket && !tooManyStreetsOnStage {
			errs = append(errs, errTooManyStreetsOnStage)
			tooManyStreetsOnStage = true
		}

		if streetsWithSections[i].SectionEnd.Value < streetsWithSections[i].SectionStart.Value {
			return errWrongSectionsInterval
		}

		sumSections += int64(len(streetsWithSections[i].Street.Value)) * (streetsWithSections[i].SectionEnd.Value - streetsWithSections[i].SectionStart.Value + 1)
	}

	if sumSections > maxSectionsPerTicket {
		errs = append(errs, errTooManySections)
	}

	return errors.Join(errs...)
}

func (h *TicketHandlerExcludeStreetFromSale) createInventTaskWithWithdrawalBySections(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var (
		errCommentBuilder strings.Builder
		performErrStages  models.PerformErrValue
		resultErrMsg      localization.LocalizedErrors
	)

	var ext excludestreetforsalemodels.ExtTicketInfoForOperationBySections

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	for i := range ext.StreetsWithSections {
		sections, err := handlerutils.GenSeqFromInterval(ext.StreetsWithSections[i].SectionStart.Value, ext.StreetsWithSections[i].SectionEnd.Value)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorStreetSectionsInvalid, nil)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errIncorrectOrderSectionInterval, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d create invent task with withdrawal by sections: %w", ticketInfo.TicketID, rejectErr)
			}
			return nil
		}

		err = h.inventTaskCreator.CreateInventTaskWithWithdrawal(ctx, createinventtaskmodels.RequestDataForCreateInventTask{
			OfficeId:   ext.OfficeID.ID,
			WhId:       ext.WhID.ID,
			Stage:      ext.StreetsWithSections[i].Stage.Value.ID,
			Streets:    ext.StreetsWithSections[i].Street.Value,
			Sections:   sections,
			Priority:   highPriorityInventTask,
			EmployeeId: ticketInfo.CreateEmployeeID,
		})
		if err != nil {
			performErrStages.Values = append(performErrStages.Values, ext.StreetsWithSections[i].Stage.Value.ID)

			if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
				h.addErrorCommentForCreateInventTask(&errCommentBuilder, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.Msg)
				h.addLocalizedErrorForCreateInventTask(&resultErrMsg, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.ErrKey, errWithMsg.MessageValues)
				logrus.Errorf("ticket %d err invent task with withdrawal %d: %v", ticketInfo.TicketID, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.Msg)
				continue
			}

			h.addErrorCommentForCreateInventTask(&errCommentBuilder, ext.StreetsWithSections[i].Stage.Value.ID, defaultErrorComment)
			h.addLocalizedErrorForCreateInventTask(&resultErrMsg, ext.StreetsWithSections[i].Stage.Value.ID, support_err_keys.KeyErrorInternal, nil)
		}
	}

	if len(performErrStages.Values) > 0 {
		performErrStages.Comment = errCommentBuilder.String()

		if len(performErrStages.Values) == len(ext.StreetsWithSections) {
			if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, performErrStages.Comment, &resultErrMsg); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			return nil
		}

		rawErrValues, err := jsoniter.Marshal(performErrStages)
		if err != nil {
			logrus.Errorf("can't marshal performErrStages for ticket %d: %v", ticketInfo.TicketID, err)
			rawErrValues, err = jsoniter.Marshal(models.PerformErrValue{Comment: defaultErrorComment})
			if err != nil {
				return fmt.Errorf("can't marshal performErrStages for ticket %d: %w", ticketInfo.TicketID, err)
			}
		}

		err = h.repo.PerformTicketWithErrComment(ctx, ticketInfo.TicketID, nil, rawErrValues, &resultErrMsg)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", ticketInfo.TicketID, err)
		}

		return nil
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerExcludeStreetFromSale) createInventTaskNoWithdrawalBySections(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var (
		errCommentBuilder strings.Builder
		performErrStages  models.PerformErrValue
		resultErrMsg      localization.LocalizedErrors
	)

	var ext excludestreetforsalemodels.ExtTicketInfoForOperationBySections

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure with err unset: %w", err)
	}

	for i := range ext.StreetsWithSections {
		sections, err := handlerutils.GenSeqFromInterval(ext.StreetsWithSections[i].SectionStart.Value, ext.StreetsWithSections[i].SectionEnd.Value)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorStreetSectionsInvalid, nil)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errIncorrectOrderSectionInterval, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d create invent task no withdrawal by sections: %w", ticketInfo.TicketID, rejectErr)
			}
			return nil
		}

		err = h.inventTaskCreator.CreateInventTaskNoWithdrawal(ctx, createinventtaskmodels.RequestDataForCreateInventTask{
			OfficeId:   ext.OfficeID.ID,
			WhId:       ext.WhID.ID,
			Stage:      ext.StreetsWithSections[i].Stage.Value.ID,
			Streets:    ext.StreetsWithSections[i].Street.Value,
			Sections:   sections,
			Priority:   highPriorityInventTask,
			EmployeeId: ticketInfo.CreateEmployeeID,
		})
		if err != nil {
			performErrStages.Values = append(performErrStages.Values, ext.StreetsWithSections[i].Stage.Value.ID)

			if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
				h.addErrorCommentForCreateInventTask(&errCommentBuilder, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.Msg)
				h.addLocalizedErrorForCreateInventTask(&resultErrMsg, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.ErrKey, errWithMsg.MessageValues)
				logrus.Errorf("ticket %d err create invent task no withdrawal by sections, stage %d: %v", ticketInfo.TicketID, ext.StreetsWithSections[i].Stage.Value.ID, errWithMsg.Msg)
				continue
			}

			h.addErrorCommentForCreateInventTask(&errCommentBuilder, ext.StreetsWithSections[i].Stage.Value.ID, defaultErrorComment)
			h.addLocalizedErrorForCreateInventTask(&resultErrMsg, ext.StreetsWithSections[i].Stage.Value.ID, support_err_keys.KeyErrorInternal, nil)
		}
	}

	if len(performErrStages.Values) > 0 {
		performErrStages.Comment = errCommentBuilder.String()

		if len(performErrStages.Values) == len(ext.StreetsWithSections) {
			if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, performErrStages.Comment, &resultErrMsg); rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}
			return nil
		}

		rawErrValues, err := jsoniter.Marshal(performErrStages)
		if err != nil {
			logrus.Errorf("can't marshal performErrStages for ticket %d: %v", ticketInfo.TicketID, err)
			rawErrValues, err = jsoniter.Marshal(models.PerformErrValue{Comment: defaultErrorComment})
			if err != nil {
				return fmt.Errorf("can't marshal performErrStages for ticket %d: %w", ticketInfo.TicketID, err)
			}
		}

		err = h.repo.PerformTicketWithErrComment(ctx, ticketInfo.TicketID, nil, rawErrValues, &resultErrMsg)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", ticketInfo.TicketID, err)
		}

		return nil
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerExcludeStreetFromSale) addErrorCommentForCreateInventTask(b *strings.Builder, stage int64, comment string) {
	if b.Len() == 0 {
		b.WriteString(startErrorCommentCreateInventTask)
	}

	b.WriteString(fmt.Sprintf("\nЭтаж: %d - Ошибка: %s", stage, comment))
}

func (h *TicketHandlerExcludeStreetFromSale) addLocalizedErrorForCreateInventTask(resultErrMsg *localization.LocalizedErrors, stage int64, keyID string, messageValues map[string]string) {
	if resultErrMsg == nil {
		resultErrMsg = &localization.LocalizedErrors{}
	}
	if len(resultErrMsg.KeysWithValues) == 0 {
		resultErrMsg.Add(support_err_keys.KeyErrorCreateInventTaskFailed, nil)
	}
	resultErrMsg.Add(
		support_err_keys.KeyFormatStageError,
		map[string]string{
			"stage_id": fmt.Sprintf("%d", stage),
		},
	)

	resultErrMsg.Add(keyID, messageValues)
}

type updaterExclusionStreets interface {
	StreetExclusionFromSaleUpdateV002(ctx context.Context, req commonhandlersmodels.HandlerRequestForStreetExclusionFromSaleUpdateV002) error
}

type inventTaskCreator interface {
	CreateInventTaskWithWithdrawal(ctx context.Context, body createinventtaskmodels.RequestDataForCreateInventTask) error
	CreateInventTaskNoWithdrawal(ctx context.Context, body createinventtaskmodels.RequestDataForCreateInventTask) error
}
