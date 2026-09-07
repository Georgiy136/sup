package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	custom_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type SupportFileManagerApiClient struct {
	c *fastclient.HttpClient
}

func NewSupportFileManagerApiClient() *SupportFileManagerApiClient {
	return &SupportFileManagerApiClient{
		c: fastclient.NewHttpClient(),
	}
}

func (f *SupportFileManagerApiClient) Configure(ctx context.Context, config configs.Config) {
	f.c.InitHttpClient("support_file_manager_api_client")(ctx, config)
}

func (f *SupportFileManagerApiClient) InitUpload(ctx context.Context, employeeID int64, body models.FileManagerInitUploadRequest) (*models.FileManagerInitUploadResponse, error) {
	const apiKey = "InitUpload"
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}
	resp, err := f.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return nil, f.c.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return nil, f.c.GenerateError(resp, fmt.Errorf("invalid response: %w", custom_errors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	switch resp.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[models.FileManagerInitUploadResponse]
		if err = jsoniter.Unmarshal(resp.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}
		return &wrapper.Data, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		customErrWrapper, parseErr := parseErrorMessageFromResponse(resp)
		if parseErr != nil {
			return nil, fmt.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}
		return nil, customErrWrapper
	default:
		return nil, f.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}
}

func (f *SupportFileManagerApiClient) ConfirmUpload(ctx context.Context, employeeID int64, body models.FileManagerConfirmUploadRequest) error {
	const apiKey = "ConfirmUpload"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	resp, err := f.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return f.c.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return f.c.GenerateError(resp, fmt.Errorf("invalid response: %w", custom_errors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		customErrWrapper, parseErr := parseErrorMessageFromResponse(resp)
		if parseErr != nil {
			return fmt.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}
		return customErrWrapper
	default:
		return f.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}
}

func (f *SupportFileManagerApiClient) GetDownloadURL(ctx context.Context, employeeID, fileID int64) (*models.FileManagerDownloadURLResponse, error) {
	const apiKey = "GetDownloadURL"
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}
	body := models.FileManagerDownloadURLRequest{FileID: fileID}
	resp, err := f.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return nil, f.c.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return nil, f.c.GenerateError(resp, fmt.Errorf("invalid response: %w", custom_errors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	switch resp.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[models.FileManagerDownloadURLResponse]
		if err = jsoniter.Unmarshal(resp.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}
		return &wrapper.Data, nil
	case http.StatusUnprocessableEntity:
		customErrWrapper, parseErr := parseErrorMessageFromResponse(resp)
		if parseErr != nil {
			return nil, fmt.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}
		return nil, customErrWrapper
	default:
		return nil, f.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}
}

func (f *SupportFileManagerApiClient) CancelUpload(ctx context.Context, employeeID int64, body models.FileManagerCancelUploadRequest) error {
	const apiKey = "CancelUpload"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	resp, err := f.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return f.c.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return f.c.GenerateError(resp, fmt.Errorf("invalid response: %w", custom_errors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		customErrWrapper, parseErr := parseErrorMessageFromResponse(resp)
		if parseErr != nil {
			return fmt.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}
		return customErrWrapper
	default:
		return f.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}
}

func (s *SupportFileManagerApiClient) AttachFilesToTicket(ctx context.Context, employeeID int64, body models.AttachFilesToTicketRequest) error {
	const apiKey = "AttachFilesToTicket"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	resp, err := s.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return s.c.GenerateError(resp, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if resp == nil {
		return s.c.GenerateError(resp, errors.New("invalid response: response is nil"), apiKey)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return s.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}

	defer fasthttp.ReleaseResponse(resp)

	return nil
}

func (s *SupportFileManagerApiClient) GetFileInfo(ctx context.Context, employeeID int64, body models.GetFileInfoRequest) (*models.GetFileInfoResponse, error) {
	const apiKey = "GetFileInfo"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	resp, err := s.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return nil, s.c.GenerateError(resp, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if resp == nil {
		return nil, s.c.GenerateError(resp, fmt.Errorf("invalid response: %w", custom_errors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	switch resp.StatusCode() {
	case http.StatusOK:
		var wrapper models.DataWrapper[models.GetFileInfoResponse]
		if err = jsoniter.Unmarshal(resp.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}
		return &wrapper.Data, nil
	case http.StatusUnprocessableEntity:
		customErrWrapper, parseErr := parseErrorMessageFromResponse(resp)
		if parseErr != nil {
			return nil, fmt.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}
		return nil, customErrWrapper
	default:
		return nil, s.c.GenerateError(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), apiKey)
	}
}

func parseErrorMessageFromResponse(response *fasthttp.Response) (*custom_errors.CustomErrorWrapper, error) {
	var detailErr custom_errors.DetailErrors
	if err := jsoniter.Unmarshal(response.Body(), &detailErr); err != nil {
		return nil, fmt.Errorf("error unmarshaling: %w", err)
	}
	if len(detailErr.Errors) == 0 {
		return nil, fmt.Errorf("no errors in response")
	}
	return &detailErr.Errors[0], nil
}
