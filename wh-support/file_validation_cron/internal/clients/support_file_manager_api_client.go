package clients

import (
	"context"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	serviceerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type supportFileManagerApiClient struct {
	cli *fastclient.HttpClient
}

func NewSupportFileManagerApiClient() *supportFileManagerApiClient {
	return &supportFileManagerApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *supportFileManagerApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("support_file_manager_api_client")(ctx, config)
}

func (s *supportFileManagerApiClient) GetUploadsForValidation() ([]models.Upload, error) {
	const apiKey = "GetUploadsForValidation"
	return s.getUploads(apiKey)
}

func (s *supportFileManagerApiClient) GetExpiredUploads() ([]models.Upload, error) {
	const apiKey = "GetExpiredUploads"
	return s.getUploads(apiKey)
}

func (s *supportFileManagerApiClient) GetUploadsForTransfer() ([]models.Upload, error) {
	const apiKey = "GetUploadsForTransfer"
	return s.getUploads(apiKey)
}

func (s *supportFileManagerApiClient) getUploads(apiKey string) ([]models.Upload, error) {
	response, err := s.cli.HTTPRequest(nil, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, s.cli.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[[]models.Upload]
		if err = jsoniter.Unmarshal(response.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return wrapper.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) GetFilesForDeletion() ([]models.File, error) {
	const apiKey = "GetFilesForDeletion"

	response, err := s.cli.HTTPRequest(nil, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, s.cli.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[[]models.File]
		if err = jsoniter.Unmarshal(response.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return wrapper.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) GetFilesNotAttachedToTicketForDeleting() ([]models.File, error) {
	const apiKey = "GetFilesNotAttachedToTicketForDeleting"

	response, err := s.cli.HTTPRequest(nil, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, s.cli.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[[]models.File]
		if err = jsoniter.Unmarshal(response.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return wrapper.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) UpdateUploadStatus(body models.ChangeStatusUploadRequest) error {
	const apiKey = "UpdateUploadStatus"

	response, err := s.cli.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) MarkDeletedFile(body models.MarkDeletedFileRequest) error {
	const apiKey = "MarkDeletedFile"

	response, err := s.cli.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) TransferFile(body models.TransferFileRequest) error {
	const apiKey = "TransferFile"

	response, err := s.cli.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *supportFileManagerApiClient) DeleteFile(body models.DeleteFileRequest) error {
	const apiKey = "DeleteFile"

	response, err := s.cli.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", serviceerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
