package auth_client

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
)

type employeeInfo struct {
	EmployeeID   int64  `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	Phone        string `json:"phone"`
	IsDeleted    bool   `json:"is_deleted"`
}

type EmployeeAuthorisationResponse struct {
	Data employeeInfo `json:"data"`
}

type AuthClient struct {
	httpClient *fastclient.HttpClient
}

func New(httpClient *fastclient.HttpClient) *AuthClient {
	return &AuthClient{httpClient: httpClient}
}

func (cli *AuthClient) Configure(ctx context.Context, cfg configs.Config) {
	cli.httpClient.InitHttpClient("auth_client")(ctx, cfg)
}

func (cli *AuthClient) SendNotificationCode(phoneNumber string) error {
	const apiKey = "send_phone_code"

	body := struct {
		PhoneNumber string `json:"phone_number"`
	}{phoneNumber}

	response, err := cli.httpClient.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return fmt.Errorf("[SendNotificationCode] error getting sending push notification, error: %w", err)
	}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode)
	}
}

func (cli *AuthClient) CheckLogin(code int64, phoneNumber string) (int64, bool, error) {
	const apiKey = "check_login"

	body := struct {
		PhoneNumber string `json:"phone_number"`
		Code        int64  `json:"code"`
	}{phoneNumber, code}

	response, err := cli.httpClient.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return 0, false, fmt.Errorf("can't send http request: %w", err)
	}

	if response == nil {
		return 0, false, common.ErrEmptyResponse
	}

	switch response.StatusCode() {
	case http.StatusOK:
		responseData := EmployeeAuthorisationResponse{}
		err = jsoniter.Unmarshal(response.Body(), &responseData)
		if err != nil {
			return 0, false, fmt.Errorf("[CheckLogin] error unmarshaling response, error: %w", err)
		}

		if responseData.Data.IsDeleted {
			return 0, false, fmt.Errorf("response data: %w", common.ErrDeletedEmployee)
		}

		return responseData.Data.EmployeeID, true, nil
	case http.StatusUnauthorized:
		return 0, false, nil
	default:
		return 0, false, fmt.Errorf("unexpected status code: %d: %w", response.StatusCode(), common.ErrWrongStatusCode)
	}
}
