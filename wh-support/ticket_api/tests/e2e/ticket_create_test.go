//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"github.com/ozontech/cute"
	"github.com/ozontech/cute/asserts/json"
	"net/http"
	"testing"
)

func Test_CreateTicket(t *testing.T) {
	setPort()

	cute.NewTestBuilder().
		Title("Создание заявки в простой категории").
		Tags("ticket").
		CreateStep("Создание заявки в простой категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(CreateTicketRequest{
				CategoryID:       3,
				Comment:          "hochu pitsu",
				CreateEmployeeID: 1,
				InfoForCreate: OfficeCreationRequest{
					Country:    "рашка",
					Address:    "улица пушкина дом колотушкина",
					IsPhysical: true,
					Date:       "23-06-2024",
					Name:       "вася",
					IsRent:     true,
					IsStorage:  true,
					Square:     123,
				},
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusOK).
		AssertBody(
			json.Present("$.data[0].ticket_id"),
		).ExecuteTest(context.Background(), t)

	//cute.NewTestBuilder(). // TODO: нет сложной категории в препроде
	//	Title("Создание заявки в сложной категории").
	//	Tags("ticket").
	//	CreateStep("Создание заявки в сложной категории").
	//	RequestBuilder(
	//		cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
	//		cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
	//		cute.WithHeadersKV("Content-Type", "application/json"),
	//		cute.WithMethod(http.MethodPost),
	//		cute.WithMarshalBody(CreateTicketRequest{
	//			CategoryID:       2,
	//			Comment:          "uzhe ne hochu pitsu",
	//			CreateEmployeeID: 1,
	//			InfoForCreate:    "{\n  \"block_id\": 123,\n  \"block_name\": \"block\",\n  \"office_name\": \"office\"\n}",
	//		}),
	//	).
	//	ExpectExecuteTimeout(timeout).
	//	ExpectStatus(http.StatusOK).
	//	AssertBody(
	//		json.Present("$.data[0].ticket_id"),
	//	).ExecuteTest(context.Background(), t)

	cute.NewTestBuilder().
		Title("Создание заявки в несуществующей категории").
		Tags("ticket").
		CreateStep("Создание заявки в несуществующей категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"), cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(CreateTicketRequest{
				CategoryID:       7,
				Comment:          "hochu pitsu",
				CreateEmployeeID: 1,
				InfoForCreate: OfficeCreationRequest{
					Country:    "рашка",
					Address:    "улица пушкина дом колотушкина",
					IsPhysical: true,
					Date:       "23-06-2024",
					Name:       "вася",
					IsRent:     true,
					IsStorage:  true,
					Square:     123,
				},
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusUnprocessableEntity).
		AssertBody(
			json.Equal("$.errors[0].error", "category.notfound"),
		).ExecuteTest(context.Background(), t)

	cute.NewTestBuilder().
		Title("Создание заявки с недостаточной информацией для категории").
		Tags("ticket").
		CreateStep("Создание заявки с недостаточной информацией для категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(CreateTicketRequest{
				CategoryID:       3,
				Comment:          "hochu pitsu",
				CreateEmployeeID: 1,
				InfoForCreate: WrongOfficeCreationRequest{
					Country:    "рашка",
					Address:    "улица пушкина дом колотушкина",
					IsPhysical: true,
					Name:       "вася",
					IsRent:     true,
					IsStorage:  true,
					Square:     123,
				},
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusUnprocessableEntity).
		AssertBody(
			json.Equal("$.errors[0].error", "categorystatus.notenoughinfo"),
		).ExecuteTest(context.Background(), t)

	cute.NewTestBuilder().
		Title("Создание заявки с категорией, по которой нельзя создать заявку").
		Tags("ticket").
		CreateStep("Создание заявки с категорией, по которой нельзя создать заявку").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(CreateTicketRequest{
				CategoryID:       2,
				Comment:          "hochu pitsu",
				CreateEmployeeID: 1,
				InfoForCreate: OfficeCreationRequest{
					Country:    "рашка",
					Address:    "улица пушкина дом колотушкина",
					IsPhysical: true,
					Date:       "23-06-2024",
					Name:       "вася",
					IsRent:     true,
					IsStorage:  true,
					Square:     123,
				},
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusUnprocessableEntity).
		AssertBody(
			json.Equal("$.errors[0].error", "category.isfornotticket"),
		).ExecuteTest(context.Background(), t)

	//cute.NewTestBuilder(). // TODO: нет таких на препроде
	//	Title("Создание заявки с категорией без следующего статуса").
	//	Tags("ticket").
	//	CreateStep("Создание заявки с категорией без следующего статуса").
	//	RequestBuilder(
	//		cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create?employee_id=123", port)),
	//		cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
	//		cute.WithHeadersKV("Content-Type", "application/json"),
	//		cute.WithMethod(http.MethodPost),
	//		cute.WithMarshalBody(CreateTicketRequest{
	//			CategoryID:       10,
	//			Comment:          "turniketi ne rabotaut(((",
	//			CreateEmployeeID: 1,
	//			InfoForCreate:    "{\n  \"block_id\": 123\n}",
	//		}),
	//	).
	//	ExpectExecuteTimeout(timeout).
	//	ExpectStatus(http.StatusUnprocessableEntity).
	//	AssertBody(
	//		json.Equal("$.errors[0].error", "categorystatus.nonextstatus"),
	//	).ExecuteTest(context.Background(), t)
}
