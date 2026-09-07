package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type ChatProviderApiClient struct {
	c *fastclient.HttpClient
}

func NewChatProviderApiClient() *ChatProviderApiClient {
	return &ChatProviderApiClient{
		c: fastclient.NewHttpClient(),
	}
}

func (c *ChatProviderApiClient) Configure(ctx context.Context, config configs.Config) {
	c.c.InitHttpClient("chat_provider_api_client")(ctx, config)
}

func (c *ChatProviderApiClient) CreateNewChat(ctx context.Context, employeeID int64, request models.CreateNewChatRequest) (*models.CreateNewChatResponse, error) {
	const apiKey = "CreateNewChat"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, request, apiKey, params)
	if err != nil {
		return nil, c.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response: %w", errResponseNil), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data models.CreateNewChatResponse `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}
		return &body.Data, nil
	default:
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (c *ChatProviderApiClient) CreatePost(ctx context.Context, employeeID int64, request models.CreatePostRequest) (*models.CreatePostResponse, error) {
	const apiKey = "CreatePost"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, request, apiKey, params)
	if err != nil {
		return nil, c.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response: %w", errResponseNil), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data models.CreatePostResponse `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return &body.Data, nil
	case http.StatusUnprocessableEntity:
		var DataError DataError

		err := jsoniter.Unmarshal(response.Body(), &DataError)
		if err != nil {
			return nil, c.c.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}

		return nil, errors.Join(DataError.ConvertToErrorSlice()...)
	case http.StatusInternalServerError:
		return nil, c.c.GenerateError(response, errors.New("internal server error"), apiKey)
	default:
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (c *ChatProviderApiClient) GetChatByChatID(ctx context.Context, employeeID int64, request models.GetChatByChatIDRequest) (*models.GetChatByChatIDResponse, error) {
	const apiKey = "GetChatByChatID"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, request, apiKey, params)
	if err != nil {
		return nil, c.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response: %w", errResponseNil), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data models.GetChatByChatIDResponse `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return &body.Data, nil
	case http.StatusUnprocessableEntity:
		var dataError DataError

		err := jsoniter.Unmarshal(response.Body(), &dataError)
		if err != nil {
			return nil, c.c.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}

		return nil, dataError
	case http.StatusInternalServerError:
		return nil, c.c.GenerateError(response, errors.New("internal server error"), apiKey)
	default:
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (c *ChatProviderApiClient) GetHistoryChatByChatID(ctx context.Context, employeeID int64, request models.GetHistoryChatByChatIDRequest) (*models.GetHistoryChatByChatIDResponse, error) {
	const apiKey = "GetHistoryChatByChatID"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.c.HttpRequestWithCtx(ctx, nil, http.MethodPost, request, apiKey, params)
	if err != nil {
		return nil, c.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response: %w", errResponseNil), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data models.GetHistoryChatByChatIDResponse `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return &body.Data, nil
	case http.StatusUnprocessableEntity:
		var DataError DataError

		err := jsoniter.Unmarshal(response.Body(), &DataError)
		if err != nil {
			return nil, c.c.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}

		return nil, errors.Join(DataError.ConvertToErrorSlice()...)
	case http.StatusInternalServerError:
		return nil, c.c.GenerateError(response, errors.New("internal server error"), apiKey)
	default:
		return nil, c.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
