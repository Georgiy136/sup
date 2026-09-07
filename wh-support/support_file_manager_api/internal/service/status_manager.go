package service

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/constant"
)

var allowedFileTransitions = map[string][]string{
	constant.FileStatusActive:        {constant.FileStatusAttachPending},
	constant.FileStatusDeletePending: {constant.FileStatusActive, constant.FileStatusAttachPending},
	constant.FileStatusDeleted:       {constant.FileStatusDeletePending},
}

var allowedUploadTransitions = map[string][]string{
	constant.UploadStatusValidating:  {constant.UploadStatusInitiated},
	constant.UploadStatusReady:       {constant.UploadStatusValidating},
	constant.UploadStatusInvalid:     {constant.UploadStatusValidating},
	constant.UploadStatusUserReject:  {constant.UploadStatusInitiated, constant.UploadStatusValidating, constant.UploadStatusReady},
	constant.UploadStatusExpired:     {constant.UploadStatusInitiated, constant.UploadStatusValidating, constant.UploadStatusReady},
	constant.UploadStatusTransferred: {constant.UploadStatusReady},
}

type StatusManager struct{}

func NewStatusManager() *StatusManager {
	return &StatusManager{}
}

func (sm *StatusManager) ValidateUploadStatusTransition(currentStatus, newStatus string) error {
	return validateTransition(currentStatus, newStatus, allowedUploadTransitions, "upload")
}

func (sm *StatusManager) CanDeleteFile(currentStatus string) (string, error) {
	if currentStatus == constant.FileStatusDeletePending {
		return "", fmt.Errorf("file is already marked for deletion")
	}
	return constant.FileStatusDeletePending, validateTransition(currentStatus, constant.FileStatusDeletePending, allowedFileTransitions, "file")
}

func (sm *StatusManager) CanAttachFile(currentStatus string) (string, error) {
	return constant.FileStatusActive, validateTransition(currentStatus, constant.FileStatusActive, allowedFileTransitions, "file")
}

func (sm *StatusManager) CanMarkFileDeleted(currentStatus string) (string, error) {
	return constant.FileStatusDeleted, validateTransition(currentStatus, constant.FileStatusDeleted, allowedFileTransitions, "file")
}

func (sm *StatusManager) CanConfirmUpload(currentStatus string) (string, error) {
	return constant.UploadStatusValidating, sm.ValidateUploadStatusTransition(currentStatus, constant.UploadStatusValidating)
}

func (sm *StatusManager) CanCancelUpload(currentStatus string) (string, error) {
	return constant.UploadStatusUserReject, sm.ValidateUploadStatusTransition(currentStatus, constant.UploadStatusUserReject)
}

func (sm *StatusManager) CanTransferFile(currentStatus string) (string, error) {
	return constant.UploadStatusTransferred, sm.ValidateUploadStatusTransition(currentStatus, constant.UploadStatusTransferred)
}

func validateTransition(currentStatus, newStatus string, allowed map[string][]string, entity string) error {
	allowedFrom, ok := allowed[newStatus]

	if !ok {
		return fmt.Errorf("unknown %s status: %s", entity, newStatus)
	}
	for _, s := range allowedFrom {
		if currentStatus == s {
			return nil
		}
	}
	return fmt.Errorf("invalid %s status transition: cannot change from %s to %s", entity, currentStatus, newStatus)
}
