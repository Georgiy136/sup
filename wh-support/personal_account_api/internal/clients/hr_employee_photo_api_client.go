package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
	"strconv"
)

type HrEmployeePhotoApiClient struct {
	c *fastclient.HttpClient
}

func NewHrEmployeePhotoApiClient(ctx context.Context, config configs.Config) *HrEmployeePhotoApiClient {
	client := &HrEmployeePhotoApiClient{
		c: fastclient.NewHttpClient(),
	}

	client.c.InitHttpClient("hr_employee_photo_api_client")(ctx, config)

	return client
}

func (s *HrEmployeePhotoApiClient) GetEmployeePhoto(employeeID int64) ([]byte, error) {
	const (
		apiKey = "GetEmployeePhoto"
		base   = 10
	)
	params := map[string]string{
		"id":    strconv.FormatInt(employeeID, base),
		"small": "true",
	}

	response, err := s.c.HTTPRequest(nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, s.c.GenerateError(response, err, apiKey)
	}

	body := struct {
		Data struct {
			Photo []byte `json:"photo_base64"`
		} `json:"data"`
	}{}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}
	default:
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}

	return body.Data.Photo, nil
}
