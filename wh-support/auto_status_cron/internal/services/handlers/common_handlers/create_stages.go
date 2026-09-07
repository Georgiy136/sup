package commonhandlers

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
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	startErrorCommentCreateStages = "Ошибка при создании этажей:"
	startErrorCommentCreateParts  = "Ошибка при создании партов для этажей:"
	defaultErrorComment           = "Внутренняя ошибка"
)

func (c *CommonHandlers) CreateStages(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateStages) error {
	var (
		errCommentBuilder strings.Builder
		resultErrMsg      localization.LocalizedErrors
		performErrStages  models.PerformErrValue
		hasInternalError  bool
	)

	errCommentsMapForCreateStages := make(map[string][]int64)

	for i := range req.Stages {
		bodyRequestForCreateStage := prepareRequestForCreateStage(req.WhID, req.OfficeID, req.Stages[i])

		bodyRequestForCreateStage.EmployeeID = req.EmployeeID

		err := c.stageApi.CreateStage(ctx, bodyRequestForCreateStage)
		if err != nil {
			performErrStages.Values = append(performErrStages.Values, req.Stages[i])

			if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
				errCommentsMapForCreateStages[errWithMsg.Msg] = append(errCommentsMapForCreateStages[errWithMsg.Msg], req.Stages[i])
				c.addLocalizationForCreateStages(&resultErrMsg, req.Stages[i], errWithMsg.ErrKey, errWithMsg.MessageValues)

				logrus.Debugf("ticket %d err create stage %d: %v", req.TicketID, req.Stages[i], errWithMsg.Msg)
				continue
			}
			hasInternalError = true

			errCommentsMapForCreateStages[defaultErrorComment] = append(errCommentsMapForCreateStages[defaultErrorComment], req.Stages[i])
			c.addLocalizationForCreateStages(&resultErrMsg, req.Stages[i], support_err_keys.KeyErrorInternal, nil)

			logrus.Debugf("ticket %d internal err create stage %d: %v", req.TicketID, req.Stages[i], err)
		}
	}

	buildErrCommentForCreateStages(&errCommentBuilder, errCommentsMapForCreateStages)
	performErrStages.Comment = errCommentBuilder.String()

	if len(performErrStages.Values) > 0 {
		if len(performErrStages.Values) == len(req.Stages) {
			// Если есть внутренняя ошибка - не отклоняем тикет, возвращаем ошибку для повторной обработки
			if hasInternalError {
				return fmt.Errorf("internal error creating stages for ticket %d", req.TicketID)
			}

			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, performErrStages.Comment, &resultErrMsg)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			return nil
		}

		rawErrValues, err := jsoniter.Marshal(performErrStages)
		if err != nil {
			logrus.Errorf("can't marshal performErrStages for ticket %d: %v", req.TicketID, err)
			rawErrValues, err = jsoniter.Marshal(models.PerformErrValue{Comment: defaultErrorComment})
			if err != nil {
				return fmt.Errorf("can't marshal performErrStages for ticket %d: %w", req.TicketID, err)
			}
		}

		err = c.repo.PerformTicketWithErrComment(ctx, req.TicketID, nil, rawErrValues, &resultErrMsg)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("can't reject ticket %d: %v", req.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", req.TicketID, err)
		}
		return nil
	}

	err := c.repo.PerformTicket(ctx, req.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", req.TicketID, err)
	}

	return nil
}

func buildErrCommentForCreateStages(b *strings.Builder, errCommentsMapForCreateStage map[string][]int64) {
	b.WriteString(startErrorCommentCreateStages)

	for comment, stages := range errCommentsMapForCreateStage {
		b.WriteString("\nЭтажи ")

		for index, stage := range stages {
			if index > 0 {
				b.WriteString(", ")
			}
			b.WriteString(strconv.FormatInt(stage, 10))
		}

		b.WriteString(fmt.Sprintf(" - %s", comment))
	}
}

func (c *CommonHandlers) addLocalizationForCreateStages(resultErrMsg *localization.LocalizedErrors, stageID int64, keyID string, messageValues map[string]string) {
	if resultErrMsg == nil {
		resultErrMsg = &localization.LocalizedErrors{}
	}
	if len(resultErrMsg.KeysWithValues) == 0 {
		resultErrMsg.Add(support_err_keys.KeyErrorCreateStagesFailed, nil)
	}
	resultErrMsg.Add(
		support_err_keys.KeyFormatStageError,
		map[string]string{
			"stage_id": fmt.Sprintf("%d", stageID),
		},
	)
	resultErrMsg.Add(keyID, messageValues)
}

func (c *CommonHandlers) CreateParts(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateParts) error {
	var (
		errCommentBuilder strings.Builder
		resultErrMsg      localization.LocalizedErrors
		performErrParts   models.PerformErrValue
		hasInternalError  bool
	)
	for i := range req.Stages {
		bodyRequestForCreatePart := prepareRequestForCreateParts(req.OfficeID, req.WhID, req.Stages[i], req.PartName)

		bodyRequestForCreatePart.EmployeeID = req.EmployeeID

		err := c.stageApi.CreatePart(ctx, bodyRequestForCreatePart)
		if err != nil {
			performErrParts.Values = append(performErrParts.Values, req.Stages[i])

			if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
				c.addRuCommentForCreateParts(&errCommentBuilder, req.Stages[i], errWithMsg.Msg)
				c.addLocalizationForCreateParts(&resultErrMsg, req.Stages[i], errWithMsg.ErrKey, errWithMsg.MessageValues)

				logrus.Debugf("ticket %d err create part for stage %d: %v", req.TicketID, req.Stages[i], errWithMsg.Msg)
				continue
			}
			hasInternalError = true

			c.addRuCommentForCreateParts(&errCommentBuilder, req.Stages[i], defaultErrorComment)
			c.addLocalizationForCreateParts(&resultErrMsg, req.Stages[i], support_err_keys.KeyErrorInternal, nil)

			logrus.Debugf("ticket %d internal err create part for stage %d: %v", req.TicketID, req.Stages[i], err)
		}
	}

	if len(performErrParts.Values) > 0 {
		performErrParts.Comment = errCommentBuilder.String()

		if len(performErrParts.Values) == len(req.Stages) {
			// Если есть внутренняя ошибка - не отклоняем тикет, возвращаем ошибку для повторной обработки
			if hasInternalError {
				return fmt.Errorf("internal error creating parts for ticket %d", req.TicketID)
			}

			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, performErrParts.Comment, &resultErrMsg)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			return nil
		}

		rawErrComment, err := jsoniter.Marshal(performErrParts)
		if err != nil {
			return fmt.Errorf("can't marshal performErrParts: %w", err)
		}

		err = c.repo.PerformTicketWithErrComment(ctx, req.TicketID, nil, rawErrComment, &resultErrMsg)
		if err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("can't reject ticket on perform %d: %v", req.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", req.TicketID, err)
		}

		return nil
	}

	err := c.repo.PerformTicket(ctx, req.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", req.TicketID, err)
	}

	return nil
}

func (c *CommonHandlers) addRuCommentForCreateParts(b *strings.Builder, stageID int64, comment string) {
	if b.Len() == 0 {
		b.WriteString(startErrorCommentCreateParts)
	}
	b.WriteString(fmt.Sprintf("\nЭтаж %d - %s", stageID, comment))
}

func (c *CommonHandlers) addLocalizationForCreateParts(resultErrMsg *localization.LocalizedErrors, stageID int64, keyID string, messageValues map[string]string) {
	if resultErrMsg == nil {
		resultErrMsg = &localization.LocalizedErrors{}
	}
	if len(resultErrMsg.KeysWithValues) == 0 {
		resultErrMsg.Add(support_err_keys.KeyErrorCreatePartsFailed, nil)
	}
	resultErrMsg.Add(
		support_err_keys.KeyFormatStageError,
		map[string]string{
			"stage_id": fmt.Sprintf("%d", stageID),
		},
	)
	resultErrMsg.Add(keyID, messageValues)
}
