//nolint:all

package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	port    = "8080"
	timeout = 5 * time.Second
)

type OfficeCreationRequest struct {
	Country    string `json:"country"`
	Address    string `json:"address"`
	IsPhysical bool   `json:"is_physical"`
	Date       string `json:"date"`
	Name       string `json:"name"`
	IsRent     bool   `json:"is_rent"`
	IsStorage  bool   `json:"is_storage"`
	Square     int    `json:"square"`
}

type WrongOfficeCreationRequest struct {
	Country    string `json:"country"`
	Address    string `json:"address"`
	IsPhysical bool   `json:"is_physical"`
	Name       string `json:"name"`
	IsRent     bool   `json:"is_rent"`
	IsStorage  bool   `json:"is_storage"`
	Square     int    `json:"square"`
}

type CreateTicketRequest struct {
	CategoryID       int64       `json:"category_id" binding:"required,gt=0"`
	Comment          string      `json:"comment" binding:"required,gt=0,lte=500"`
	CreateEmployeeID int64       `json:"create_employee_id" binding:"required,EmployeeID"`
	InfoForCreate    interface{} `json:"info_for_create" binding:"required"`
}

type CreateTicketResponseData struct {
	TicketID int64 `json:"ticket_id"`
}

type CreateTicketResponse struct {
	Data []CreateTicketResponseData `json:"data"`
}

type RejectTicketResponseData struct {
	TicketID []int64 `json:"ticket_id"`
}

type RejectTicketResponse struct {
	Data []RejectTicketResponseData `json:"data"`
}

type GetCategoryInfoRequest struct {
	StatusID string `json:"status_id" binding:"omitempty,gt=0,lte=3"`
}

type InfoModelData struct {
	DataName              string `json:"data_name"`
	DataType              string `json:"data_type"`
	IsRequired            bool   `json:"is_required"`
	FrontDataName         string `json:"front_data_name"`
	FrontDataType         string `json:"front_data_type"`
	FrontDataDescription  string `json:"front_data_description"`
	FrontMaxThresholdData string `json:"front_max_threshold_data"`
	FrontMinThresholdData string `json:"front_min_threshold_data"`
}

type GetCategoryInfoResponseData struct {
	Priority              int             `json:"priority"`
	IsParent              bool            `json:"is_parent"`
	StatusId              string          `json:"status_id"`
	CategoryId            int             `json:"category_id"`
	CategoryName          string          `json:"category_name"`
	NextStatusId          string          `json:"next_status_id"`
	ApproveGroupId        int             `json:"approve_group_id"`
	PerformGroupId        int             `json:"perform_group_id"`
	InformationModel      []InfoModelData `json:"information_model"`
	ApproveGroupName      string          `json:"approve_group_name"`
	PerformGroupName      string          `json:"perform_group_name"`
	StatusDescription     string          `json:"status_description"`
	IdealPerformTimeSec   int             `json:"ideal_perform_time_sec"`
	NextStatusDescription string          `json:"next_status_description"`
}

type GetCategoryInfoResponse struct {
	Data []GetCategoryInfoResponseData `json:"data"`
}

type TreeCategory struct {
	Children         []TreeCategory `json:"children"`
	CategoryId       int            `json:"category_id"`
	CategoryName     string         `json:"category_name"`
	ParentCategoryId int            `json:"parent_category_id"`
}

type GetCategoryTreeResponse struct {
	Data []TreeCategory `json:"data"`
}

type AssertBody func(body []byte) error

func customAssertRejectBody(queries map[string][]string) AssertBody {
	return func(bytes []byte) error {
		if len(bytes) == 0 {
			return errors.New("response body is empty")
		}

		var resp RejectTicketResponse

		err := json.Unmarshal(bytes, &resp)
		if err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		if fmt.Sprintf("%d", resp.Data[0].TicketID[0]) != queries["ticket_id"][0] {
			return errors.New(fmt.Sprintf("wrong response expect %s got %d", queries["ticket_id"][0], resp.Data[0].TicketID[0]))
		}

		return nil
	}
}

func customAssertCategoryInfoBody() AssertBody {
	return func(bytes []byte) error {
		if len(bytes) == 0 {
			return errors.New("response body is empty")
		}

		var resp GetCategoryInfoResponse

		err := json.Unmarshal(bytes, &resp)
		if err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		expect1 := "country"
		expect2 := "address"
		got1 := resp.Data[0].InformationModel[0].DataName
		got2 := resp.Data[0].InformationModel[1].DataName

		if got1 != expect1 || got2 != expect2 {
			return errors.New(fmt.Sprintf("wrong info models: expect \n%s\n%s\n got \n%s\n%s\n", expect1, expect2, got1, got2))
		}

		return nil
	}
}

func customAssertParseCategoryInfoBody() AssertBody {
	return func(bytes []byte) error {
		if len(bytes) == 0 {
			return errors.New("response body is empty")
		}

		var resp GetCategoryInfoResponse

		err := json.Unmarshal(bytes, &resp)
		if err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		if resp.Data[0].CategoryId != 3 {
			return fmt.Errorf("wrong category data: expect 1 got %d", resp.Data[0].CategoryId)
		}

		return nil
	}
}

func findCategoryInTree(n int, tree []TreeCategory) bool {
	if len(tree) == 0 {
		return false
	}

	for _, v := range tree {
		if v.CategoryId == n {
			return true
		}
	}

	for _, v := range tree {
		if len(v.Children) != 0 {
			if findCategoryInTree(n, v.Children) {
				return true
			}
		}
	}

	return false
}

func customAssertTreeBody() AssertBody {
	return func(bytes []byte) error {
		if len(bytes) == 0 {
			return errors.New("response body is empty")
		}

		var resp GetCategoryTreeResponse

		err := json.Unmarshal(bytes, &resp)
		if err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		if !findCategoryInTree(1, resp.Data) {
			return errors.New("no category #1 in tree")
		}

		return nil
	}
}

func getCreatedTicketId(response *http.Response) (int64, error) {
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("cannot read response body: %w", err)
	}

	var resp CreateTicketResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return 0, fmt.Errorf("cannot unmarshal body: %w", err)
	}

	cur := resp.Data[0].TicketID
	for i := range resp.Data {
		if resp.Data[i].TicketID < cur {
			cur = resp.Data[i].TicketID
		}
	}

	return cur, nil
}

func setPort() {
	portEnv := os.Getenv("PORT")
	if portEnv != "" {
		port = portEnv
	}
}

func setQueries() map[string][]string {
	queries := map[string][]string{"ticket_id": nil}
	queries["ticket_id"] = make([]string, 1)
	queries["employee_id"] = []string{"123"}

	return queries
}
