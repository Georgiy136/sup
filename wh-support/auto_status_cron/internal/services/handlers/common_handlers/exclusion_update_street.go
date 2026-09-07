package commonhandlers

import (
	"context"
	"errors"
	"fmt"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	"github.com/sirupsen/logrus"
)

const (
	CommentErrBadRequestSectionOnStreets = "Ошибка составления заявки: Проверьте правильность введенных секций для улиц"
)

func (c *CommonHandlers) ExclusionUpdateStreet(ctx context.Context, req commonhandlersmodels.HandlerRequestForExclusionUpdateStreet) error {
	bodyRequest := prepareRequestDataForExcludeStreetForAssembly(req)

	err := c.assemblyApi.AssemblyStreetExclusionUpdate(ctx, bodyRequest)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on update exlusion street from asseembly: %v", req.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't update exlusion street from asseembly: %w", err)
	}

	err = c.repo.PerformTicket(ctx, req.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d on update exlusion street from asseembly: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", req.TicketID, err)
	}

	return nil
}

func (c *CommonHandlers) StreetExclusionFromSaleUpdateV002(ctx context.Context, req commonhandlersmodels.HandlerRequestForStreetExclusionFromSaleUpdateV002) error {
	bodyRequest := prepareRequestDataForStreetExclusionFromSaleUpdateV002(req)

	if err := c.validator.Struct(bodyRequest); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorStreetSectionsInvalid, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, CommentErrBadRequestSectionOnStreets, locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't ticket %d reject on update exclusion street from sale: %v", req.TicketID, rejectErr)
		}

		return nil
	}

	err := c.assemblyApi.StreetExclusionFromSaleUpdateV002(ctx, bodyRequest)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on update exclusion street from sale: %v", req.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't update exclusion street from sale: %w", err)
	}

	err = c.repo.PerformTicket(ctx, req.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("ticket %d reject on update exclusion street from sale: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", req.TicketID, err)
	}

	return nil
}

func (c *CommonHandlers) HandleWhOrStagesExclusionFromAssembly(ctx context.Context, req commonhandlersmodels.HandlerRequestForStageOrWhExclusionFromAssemblyUpdateV001) error {
	bodyRequest := prepareRequestDataForStageOrWhExclusionFromAssemblyUpdateV001(req)

	err := c.assemblyApi.StageOrWhExclusionFromAssemblyUpdateV001(ctx, bodyRequest)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := c.repo.RejectTicket(ctx, req.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", req.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject on stage or wh exclusion from assembly: %v", req.TicketID, errWithMsg.Msg)
			return nil
		}

		return fmt.Errorf("can't update stage or wh exclusion from assembly: %w", err)
	}

	if err = c.repo.PerformTicket(ctx, req.TicketID, nil); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := c.repo.RejectTicket(ctx, req.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d on stage or wh exclusion from assembly: %v", req.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", req.TicketID, err)
	}

	return nil
}
