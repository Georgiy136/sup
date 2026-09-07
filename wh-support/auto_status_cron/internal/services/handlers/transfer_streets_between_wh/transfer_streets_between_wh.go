package transferstreetsbetweenwh

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
	transferstreetsbetweenwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/transfer_streets_between_wh/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerTransferStreetsBetweenWh"

	startErrorCommentTransferStreets = "Ошибка при переносе улиц между блоками:"
	defaultErrorComment              = "Внутренняя ошибка"
)

type TicketHandlerTransferStreetsBetweenWh struct {
	repo                  services.HandlerTicketsRepo
	transferWhForStockApi transferWhForStockApi
}

func NewTicketHandlerTransferStreetsBetweenWh(repo services.HandlerTicketsRepo, transferWhForStockApi transferWhForStockApi) *TicketHandlerTransferStreetsBetweenWh {
	return &TicketHandlerTransferStreetsBetweenWh{
		repo:                  repo,
		transferWhForStockApi: transferWhForStockApi,
	}
}

func (t *TicketHandlerTransferStreetsBetweenWh) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var (
		ext transferstreetsbetweenwhmodels.ExtTicketInfoForTransferStreetsBetweenWh

		errCommentBuilder strings.Builder
		resultErrMsg      localization.LocalizedErrors
		performErr        models.PerformErrValue
		hasInternalError  bool
	)

	errCommentsMapForTransferStreets := make(map[string][]transferstreetsbetweenwhmodels.ErrCommentValue)

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	for index := range ext.StreetsSections {
		request, err := convertExtInfoToRequestForTransferWhForStock(ext, ticketInfo.CreateEmployeeID, int64(index))
		if err != nil {
			return fmt.Errorf("[%s] can't convert ext info to request: %w", handlerName, err)
		}

		if err = t.transferWhForStockApi.TransferWhForStock(ctx, request); err != nil {
			errCommentValue := convertRequestToErrCommentValue(request)
			performErr.Values = append(performErr.Values, errCommentValue)
			if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
				errCommentsMapForTransferStreets[errWithMsg.Msg] = append(errCommentsMapForTransferStreets[errWithMsg.Msg], errCommentValue)
				addLocalizationForTransferStreets(&resultErrMsg, errCommentValue, errWithMsg.ErrKey, errWithMsg.MessageValues)
				logrus.Errorf("[%s] ticket %d: err transfer streets on stage %d: %v", handlerName, ticketInfo.TicketID, request.Stage, errWithMsg.Msg)
				continue
			}

			hasInternalError = true
			errCommentsMapForTransferStreets[defaultErrorComment] = append(errCommentsMapForTransferStreets[defaultErrorComment], errCommentValue)
			addLocalizationForTransferStreets(&resultErrMsg, errCommentValue, support_err_keys.KeyErrorInternal, nil)
			logrus.Errorf("[%s] ticket %d: internal err transfer streets on stage %d: %v", handlerName, ticketInfo.TicketID, request.Stage, err)
		}
	}

	buildErrCommentForTransferStreets(&errCommentBuilder, errCommentsMapForTransferStreets)
	performErr.Comment = errCommentBuilder.String()

	if len(performErr.Values) > 0 {
		if len(performErr.Values) == len(ext.StreetsSections) {
			if hasInternalError {
				return fmt.Errorf("[%s] internal error transfering streets for ticket %d", handlerName, ticketInfo.TicketID)
			}
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, performErr.Comment, &resultErrMsg); rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}
			return nil
		}

		rawErrValues, err := jsoniter.Marshal(performErr)
		if err != nil {
			logrus.Errorf("[%s] can't marshal performErr for ticket %d: %v", handlerName, ticketInfo.TicketID, err)
			rawErrValues, err = jsoniter.Marshal(models.PerformErrValue{Comment: defaultErrorComment})
			if err != nil {
				return fmt.Errorf("[%s] can't marshal performErr for ticket %d: %w", handlerName, ticketInfo.TicketID, err)
			}
		}

		if err = t.repo.PerformTicketWithErrComment(ctx, ticketInfo.TicketID, nil, rawErrValues, &resultErrMsg); err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
				logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
			}
			return fmt.Errorf("[%s] can't perform ticket with err comment %d: %w", handlerName, ticketInfo.TicketID, err)
		}
		return nil
	}

	if err := t.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}
		return fmt.Errorf("[%s] can't perform ticket %d: %w", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

func addLocalizationForTransferStreets(resultErrMsg *localization.LocalizedErrors, errValue transferstreetsbetweenwhmodels.ErrCommentValue, keyID string, messageValues map[string]string) {
	if resultErrMsg == nil {
		resultErrMsg = &localization.LocalizedErrors{}
	}
	if len(resultErrMsg.KeysWithValues) == 0 {
		resultErrMsg.Add(support_err_keys.KeyErrorStreetTransferFailed, nil)
	}
	resultErrMsg.Add(
		support_err_keys.KeyFormatStageWithStreetsAndSections,
		map[string]string{
			"stage_id":      strconv.FormatInt(errValue.Stage, 10),
			"street_start":  strconv.FormatInt(errValue.StreetStart, 10),
			"street_end":    strconv.FormatInt(errValue.StreetEnd, 10),
			"section_start": strconv.FormatInt(errValue.SectionStart, 10),
			"section_end":   strconv.FormatInt(errValue.SectionEnd, 10),
		},
	)
	resultErrMsg.Add(keyID, messageValues)
}

func buildErrCommentForTransferStreets(b *strings.Builder, errMap map[string][]transferstreetsbetweenwhmodels.ErrCommentValue) {
	b.WriteString(startErrorCommentTransferStreets)

	for comment, streetsSections := range errMap {
		for index, val := range streetsSections {
			if index > 0 {
				b.WriteString("; ")
			}
			b.WriteString(fmt.Sprintf("\nЭтаж %d, ряды %d - %d, секции %d - %d", val.Stage, val.StreetStart, val.StreetEnd, val.SectionStart, val.SectionEnd))
		}

		b.WriteString(fmt.Sprintf(" - %s", comment))
	}
}

type transferWhForStockApi interface {
	TransferWhForStock(ctx context.Context, body transferstreetsbetweenwhmodels.RequestForTransferWhForStock) error
}
