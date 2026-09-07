package sprut

import (
	"context"
	"errors"
	"fmt"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	createwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_wh/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type PositiveResponseV001 struct {
	ApiCallBacks ApiCallbacks `json:"api_callback"`
}

type ApiCallbacks struct {
	Apis    []Apis  `json:"apis"`
	Token   *string `json:"token"`
	Version string  `json:"version"`
}

type Apis struct {
	Key  string   `json:"key"`
	Urls []string `json:"urls"`
}

type SprutAccessClients struct {
	sprutBackendApiClient *WhBackendApiClient
	sprutApiClient        *SprutApiClient
}

func NewSprutAccessClients() *SprutAccessClients {
	return &SprutAccessClients{
		sprutBackendApiClient: NewWhBackendApiClient(),
		sprutApiClient:        NewSprutClient(),
	}
}

func (s *SprutAccessClients) Configure(ctx context.Context, config configs.Config) {
	s.sprutBackendApiClient.Configure(ctx, config)
	s.sprutApiClient.Configure(ctx, config)
}

func (s *SprutAccessClients) AddNewBuilding(ctx context.Context, body createwhmodels.RequestDataForAddNewBuilding) error {
	var (
		callBackApiKey = "WhBuildingsAdministrationGet"
		sprutApiKey    = "wh.buildings.administration.building.update/v001"
	)

	return s.executeSprutRequest(ctx, body, callBackApiKey, sprutApiKey)
}

func (s *SprutAccessClients) executeSprutRequest(ctx context.Context, body any, callBackApiKey, sprutApiKey string) error {
	rawBody, err := jsoniter.Marshal(body)
	if err != nil {
		return fmt.Errorf("can't marshal json req body: %w", err)
	}

	respWithCallbacks, err := s.sprutBackendApiClient.GetCallBacks(ctx, callBackApiKey, rawBody)
	if err != nil {
		return fmt.Errorf("can't get callbacks: %w", err)
	}

	if respWithCallbacks == nil {
		return fmt.Errorf("response with callback is empty: %w", err)
	}

	token := respWithCallbacks.ApiCallBacks.Token
	if token == nil {
		return fmt.Errorf("sprut token empty: %w", err)
	}

	urls, err := getUrls(respWithCallbacks.ApiCallBacks.Apis, sprutApiKey)
	if err != nil {
		return fmt.Errorf("can't get urls: %w", err)
	}

	sprutApiResponse, err := s.sprutApiClient.CallSprutAPIByUrls(ctx, *token, urls, rawBody)
	if err != nil {
		if !errors.Is(err, internalerrors.ErrUnprocessableEntity) {
			return fmt.Errorf("can't call sprut api: %w", err)
		}

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(sprutApiResponse)
		if parseErr != nil {
			logrus.Errorf("can't parse response error message: %v", parseErr)
		}

		return errWithMsg
	}

	return nil
}

func getUrls(apis []Apis, apiKeySprut string) ([]string, error) {
	for _, api := range apis {
		if api.Key == apiKeySprut {
			return api.Urls, nil
		}
	}
	return nil, fmt.Errorf("can't find api with key [%s]: %w", apiKeySprut, internalerrors.ErrUrlsEmpty)
}
