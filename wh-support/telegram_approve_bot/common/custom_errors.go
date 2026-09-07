package common

import (
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	ErrNilMsg                     = errors.New("nil message was passed")
	ErrNorMsgNeitherCallback      = errors.New("update was nor msg neither callback")
	ErrInvalid8PhoneNumber        = errors.New("phone number with a 8-xxx-xxx-xx-xx format should have a length of 11 symbols")
	ErrInvalidPlus7PhoneNumber    = errors.New("phone number with a +7-xxx-xxx-xx-xx format should have a length of 12 symbols")
	ErrInvalid7PhoneNumber        = errors.New("phone number with a 7-xxx-xxx-xx-xx format should have a length of 11 symbols")
	ErrUndefinedPhoneNumberFormat = errors.New("undefined phone number format")
	ErrContactMessage             = errors.New("wrong contact message format")
	ErrCheckLogin                 = errors.New("access client couldn't check code")
	ErrDeletedEmployee            = errors.New("employee deleted")
	ErrSendNotification           = errors.New("can't send notification")
	ErrWrongStatusCode            = errors.New("wrong status code")
	ErrUnknownUpdateType          = errors.New("unknown update type")
	ErrExpectedNilTgMessage       = errors.New("expected nil telegram message")
	ErrCommandNotExist            = errors.New("command not exist")
	ErrUnknownState               = errors.New("unknown state")
	ErrInvalidState               = errors.New("invalid state")
	ErrChatAlreadyExists          = errors.New("tgchat.alreadyexists")
	ErrGetDataFromDB              = errors.New("can't get data from db")
	ErrUserAlreadyExists          = errors.New("tguserchat.alreadyexists")
	ErrHasNoRights                = errors.New("tguserchat.employeehasnorights")
	ErrUserAlreadyRegistered      = errors.New("user already registered")
	ErrUserCannotRegistered       = errors.New("user cannot be registered")
	ErrPhoneNumberNil             = errors.New("phone number nil")
	ErrInvalidType                = errors.New("invalid type")
	ErrEmptyResponse              = errors.New("empty response")
	ErrUnknown                    = errors.New("unknown")
	ErrEmployeeUnknown            = errors.New("can't find employee")
	ErrUnknownAction              = errors.New("unknown action")
	ErrEmptyCallbackMessage       = errors.New("empty callback message")
	ErrEmptyData                  = errors.New("no data in db")
	ErrSepNotFound                = errors.New("sep not found")
	ErrNilValue                   = errors.New("nil value")
	ErrWrongTicketModel           = errors.New("tickets.addedmodelerror")
	ErrTicketNotNeedApprove       = errors.New("tickets.notneedinapprove")
	ErrTicketNotFound             = errors.New("tickets.notfound")
)

type DBExecError struct {
	rest_data.CustomError
}

func (e *DBExecError) Error() string {
	return fmt.Sprintf("code: %s, msg: %s", e.CustomError.ErrorKey, e.CustomError.Message)
}
