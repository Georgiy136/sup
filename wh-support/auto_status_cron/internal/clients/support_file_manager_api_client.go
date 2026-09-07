package clients

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
)

type SupportFileManagerApiClient struct {
	cli *fastclient.HttpClient
}

func NewSupportFileManagerApiClient() *SupportFileManagerApiClient {
	return &SupportFileManagerApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (c *SupportFileManagerApiClient) Configure(ctx context.Context, config configs.Config) {
	c.cli.InitHttpClient("support_file_manager_api_client")(ctx, config)
}

func (c *SupportFileManagerApiClient) UploadFile(ctx context.Context, fileBytes []byte, fileName string) (fileID int64, err error) {
	const apiKey = "UploadFile"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	if err = writer.WriteField("entity_type", "TICKET"); err != nil {
		return 0, fmt.Errorf("failed to write entity_type field: %w", err)
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	h.Set("Content-Type", mimeType)

	fw, err := writer.CreatePart(h)
	if err != nil {
		return 0, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = fw.Write(fileBytes); err != nil {
		return 0, fmt.Errorf("failed to write file to form: %w", err)
	}

	contentType := writer.FormDataContentType()
	if err = writer.Close(); err != nil {
		return 0, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	opts := []fastclient.Opts{
		func(request *fasthttp.Request) error {
			request.Header.SetContentType(contentType)
			return nil
		},
	}

	response, err = c.cli.HTTPRequestWithOpts(nil, http.MethodPost, buf.Bytes(), apiKey, nil, opts...)
	if err != nil {
		return 0, c.cli.GenerateError(response, fmt.Errorf("upload request failed: %w", err), apiKey)
	}
	if response == nil {
		return 0, c.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	if response.StatusCode() != http.StatusOK {
		return 0, c.cli.GenerateError(response, fmt.Errorf("upload returned status %d", response.StatusCode()), apiKey)
	}

	var result models.DataWrapper[struct {
		FileID int64 `json:"file_id"`
	}]
	if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse upload response: %w", err)
	}

	if result.Data.FileID == 0 {
		return 0, fmt.Errorf("upload response missing file_id")
	}

	return result.Data.FileID, nil
}
