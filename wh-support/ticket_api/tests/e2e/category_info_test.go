//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"github.com/ozontech/cute"
	cute_json "github.com/ozontech/cute/asserts/json"
	"net/http"
	"testing"
)

func Test_InfoCreateRejectTicket(t *testing.T) {
	setPort()

	queries := setQueries()

	cute.NewTestBuilder().
		Title("Получение информации с указанием статуса -> создание -> отклонение заявки").
		Tags("category").
		CreateStep("Получение информации о заявке с указанием статуса").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/category/info", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
			cute.WithQueryKV("category_id", "3"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(GetCategoryInfoRequest{
				StatusID: "AOS",
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cute.AssertBody(customAssertParseCategoryInfoBody()),
		).NextTest().
		CreateStep("Создание заявки в простой категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
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
			cute_json.Present("$.data[0].ticket_id"),
		).NextTest().
		AfterTestExecute(
			func(response *http.Response, errors []error) error {
				ticketID, err := getCreatedTicketId(response)
				if err != nil {
					return fmt.Errorf("cannot get created ticket id: %w", err)
				}

				queries["ticket_id"][0] = fmt.Sprintf("%d", ticketID)

				return nil
			},
		).CreateStep("Отклонение заявки в простой категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/reject", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodGet),
			cute.WithQuery(queries),
		).ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusOK).AssertBody(
		cute.AssertBody(customAssertRejectBody(queries)),
	).ExecuteTest(context.Background(), t)

	cute.NewTestBuilder().
		Title("Получение информации без указания статуса -> создание -> отклонение заявки").
		Tags("category").
		CreateStep("Получение информации о заявке без указания статуса").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/category/info", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
			cute.WithQueryKV("category_id", "3"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(GetCategoryInfoRequest{
				StatusID: "",
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cute.AssertBody(customAssertCategoryInfoBody()),
		).NextTest().
		CreateStep("Создание заявки в простой категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
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
			cute_json.Present("$.data[0].ticket_id"),
		).NextTest().
		AfterTestExecute(
			func(response *http.Response, errors []error) error {
				ticketID, err := getCreatedTicketId(response)
				if err != nil {
					return fmt.Errorf("cannot get created ticket id: %w", err)
				}

				queries["ticket_id"][0] = fmt.Sprintf("%d", ticketID)

				return nil
			},
		).CreateStep("Отклонение заявки в простой категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/reject", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodGet),
			cute.WithQuery(queries),
		).ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusOK).AssertBody(
		cute.AssertBody(customAssertRejectBody(queries)),
	).ExecuteTest(context.Background(), t)

	cute.NewTestBuilder().
		Title("Получение информации о несуществующей категории -> создание -> отклонение заявки").
		Tags("ticket").
		CreateStep("Получение информации о заявке с указанием статуса").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/category/info", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
			cute.WithQueryKV("category_id", "3"),
			cute.WithMethod(http.MethodPost),
			cute.WithMarshalBody(GetCategoryInfoRequest{
				StatusID: "KEK",
			}),
		).
		ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusUnprocessableEntity).
		AssertBody(
			cute_json.Equal("$.errors[0].error", "categorystatus.notfound"),
		).NextTest().
		CreateStep("Создание заявки в несуществующей категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/create", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithQueryKV("employee_id", "123"),
			cute.WithMethod(http.MethodPost),
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
			cute_json.Equal("$.errors[0].error", "category.notfound"),
		).NextTest().CreateStep("Отклонение заявки в несуществующей категории").
		RequestBuilder(
			cute.WithURI(fmt.Sprintf("http://localhost:%s/api/ticket/reject", port)),
			cute.WithHeadersKV("Authorization", "Basic YWRtaW46dGVzdA=="),
			cute.WithHeadersKV("Content-Type", "application/json"),
			cute.WithMethod(http.MethodGet),
			cute.WithQueryKV("employee_id", "123"),
			cute.WithQueryKV("ticket_id", queries["ticket_id"][0]),
		).ExpectExecuteTimeout(timeout).
		ExpectStatus(http.StatusUnprocessableEntity).AssertBody(
		cute_json.Equal("$.errors[0].error", "tickets.notfound"),
	).ExecuteTest(context.Background(), t)
}
