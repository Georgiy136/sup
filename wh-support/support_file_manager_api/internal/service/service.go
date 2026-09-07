package service

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

type SupportFileManagerService struct {
	s3Client         S3ClientInterface
	centrifugoClient CentrifugoClientInterface
	fileValidator    FileValidator
	repository       Repository
	statusManager    *StatusManager
	errBuilder       core_errors.ErrorBuilder
	urlExpiration    models.PresignedURLSettings
}

func NewSupportFileManagerService(
	s3Client S3ClientInterface,
	centrifugoClient CentrifugoClientInterface,
	fileValidator FileValidator,
	repository Repository) *SupportFileManagerService {
	errorsRepo := errors_keys.NewErrorsRepository()
	errorsRepo.SetErrors(support_err_keys.ErrorsRepository)

	return &SupportFileManagerService{
		s3Client:         s3Client,
		centrifugoClient: centrifugoClient,
		fileValidator:    fileValidator,
		repository:       repository,
		statusManager:    NewStatusManager(),
		errBuilder:       core_errors.NewErrorBuilder(errorsRepo.GetErrors(), nil),
	}
}

func (s *SupportFileManagerService) Configure(ctx context.Context, config configs.Config) {
	const cfgKey = "presigned_url_settings"

	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired(cfgKey), &s.urlExpiration); err != nil {
		logrus.Panicf("error unmarshaling presigned url settings: %v", err)
	}
}
