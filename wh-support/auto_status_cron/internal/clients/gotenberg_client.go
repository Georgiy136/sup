package clients

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/valyala/fasthttp"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
)

type GotenbergClient struct {
	cli *fastclient.HttpClient
}

func NewGotenbergClient() *GotenbergClient {
	return &GotenbergClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (c *GotenbergClient) Configure(ctx context.Context, cfg configs.Config) {
	c.cli.InitHttpClient("gotenberg_client")(ctx, cfg)
}

func (c *GotenbergClient) ConvertRenderTemplateToPdf(ctx context.Context, docxBytes []byte) (result []byte, err error) {
	const apiKey = "ConvertDocxToPdf"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("file", "document.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = fw.Write(docxBytes); err != nil {
		return nil, fmt.Errorf("failed to write docx to form: %w", err)
	}

	contentType := writer.FormDataContentType()
	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	opts := []fastclient.Opts{
		func(request *fasthttp.Request) error {
			request.Header.SetContentType(contentType)
			return nil
		},
	}

	response, err = c.cli.HTTPRequestWithOpts(nil, http.MethodPost, body.Bytes(), apiKey, nil, opts...)
	if err != nil {
		return nil, c.cli.GenerateError(response, fmt.Errorf("gotenberg request failed: %w", err), apiKey)
	}
	if response == nil {
		return nil, c.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	if response.StatusCode() != http.StatusOK {
		return nil, c.cli.GenerateError(response, fmt.Errorf("gotenberg returned unexpected status %d", response.StatusCode()), apiKey)
	}

	pdfBytes := make([]byte, len(response.Body()))
	copy(pdfBytes, response.Body())

	return pdfBytes, nil
}
