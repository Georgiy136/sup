package clients

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/models"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type SprutAccessService struct {
	backendApiClient *WhBackendApiClient
	sprutClient      *SprutApiClient
}

func NewSprutAccessService(backendApiClient *WhBackendApiClient, sprutClient *SprutApiClient) *SprutAccessService {
	return &SprutAccessService{
		backendApiClient: backendApiClient,
		sprutClient:      sprutClient,
	}
}

func (a *SprutAccessService) Configure(ctx context.Context, cfg configs.Config) {
	a.backendApiClient.Configure(ctx, cfg)
	a.sprutClient.Configure(ctx, cfg)
}

func (a *SprutAccessService) GetBuildingsByOfficeID(ctx *gin.Context, officeID int64, employeeID int64) ([]models.BuildingInfo, error) {
	const (
		sprutApiKey   = "wh.buildings.administration.building.get.by_office/v001"
		backendApiKey = "BuildingsAdministrationGetApi"
	)

	dataForGetCallBacks := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
	}{officeID, employeeID}

	dataForGetCallBacksToSend, err := jsoniter.Marshal(dataForGetCallBacks)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for get call backs: %w", err)
	}

	respWithCallbacks, err := a.backendApiClient.GetCallBacks(backendApiKey, dataForGetCallBacksToSend)
	if err != nil {
		return nil, fmt.Errorf("can't get callbacks: %w", err)
	}

	if respWithCallbacks == nil {
		return nil, fmt.Errorf("response with callback is empty: %w", err)
	}

	token := respWithCallbacks.ApiCallBacks.Token
	if token == nil {
		return nil, fmt.Errorf("sprut token empty: %w", err)
	}

	urls, err := getUrls(respWithCallbacks.ApiCallBacks.Apis, sprutApiKey)
	if err != nil {
		return nil, fmt.Errorf("can't get urls: %w", err)
	}

	dataForSprut := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
	}{officeID, employeeID}

	dataForSprutToSend, err := jsoniter.Marshal(dataForSprut)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for sprut call backs: %w", err)
	}

	sprutApiResponse, err := a.sprutClient.CallSprutAPIByUrls(ctx, *token, urls, dataForSprutToSend)
	if err != nil {
		return nil, fmt.Errorf("can't call sprut api: %w", err)
	}

	if sprutApiResponse == nil {
		return nil, nil
	}

	var buildingsResponse models.BuildingsResponse

	err = jsoniter.Unmarshal(sprutApiResponse, &buildingsResponse)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal sprut response: %w", err)
	}

	return buildingsResponse.Data, nil
}

func (a *SprutAccessService) GetOfficeStagePartByWhID(ctx *gin.Context, employeeID, officeID, whID int64) (*models.StagesInfo, error) {
	const (
		sprutApiKey   = "wh.sprutenv.office.stagepart.get.by_wh/v001"
		backendApiKey = "WhSprutenvAdministrationGetApi"
	)

	dataForGetCallBacks := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
	}{officeID, employeeID}

	dataForGetCallBacksToSend, err := jsoniter.Marshal(dataForGetCallBacks)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for get call backs: %w", err)
	}

	respWithCallbacks, err := a.backendApiClient.GetCallBacks(backendApiKey, dataForGetCallBacksToSend)
	if err != nil {
		return nil, fmt.Errorf("can't get callbacks: %w", err)
	}

	if respWithCallbacks == nil {
		return nil, fmt.Errorf("response with callback is empty: %w", err)
	}

	token := respWithCallbacks.ApiCallBacks.Token
	if token == nil {
		return nil, fmt.Errorf("sprut token empty: %w", err)
	}

	urls, err := getUrls(respWithCallbacks.ApiCallBacks.Apis, sprutApiKey)
	if err != nil {
		return nil, fmt.Errorf("can't get urls: %w", err)
	}

	dataForSprut := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
		WhID       int64 `json:"wh_id"`
	}{officeID, employeeID, whID}

	dataForSprutToSend, err := jsoniter.Marshal(dataForSprut)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for sprut call backs: %w", err)
	}

	sprutApiResponse, err := a.sprutClient.CallSprutAPIByUrls(ctx, *token, urls, dataForSprutToSend)
	if err != nil {
		return nil, fmt.Errorf("can't call sprut api: %w", err)
	}

	if sprutApiResponse == nil {
		return nil, nil
	}

	var stagesResponse models.StagesResponse

	err = jsoniter.Unmarshal(sprutApiResponse, &stagesResponse)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal sprut response: %w", err)
	}

	return &stagesResponse.Data, nil
}

func (a *SprutAccessService) GetStoragePlacesBySections(ctx *gin.Context, employeeID, officeID, whID, street int64, stage *int64, sectionStart, sectionEnd int64) ([]models.StoragePlace, error) {
	const (
		sprutApiKey   = "sprutenv.storageplaces.info.get.by_sections/external/v001"
		backendApiKey = "WhSprutenvAdministrationGetApi"
	)

	dataForGetCallBacks := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
	}{officeID, employeeID}

	dataForGetCallBacksToSend, err := jsoniter.Marshal(dataForGetCallBacks)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for get call backs: %w", err)
	}

	respWithCallbacks, err := a.backendApiClient.GetCallBacks(backendApiKey, dataForGetCallBacksToSend)
	if err != nil {
		return nil, fmt.Errorf("can't get callbacks: %w", err)
	}

	if respWithCallbacks == nil {
		return nil, fmt.Errorf("response with callback is empty: %w", err)
	}

	token := respWithCallbacks.ApiCallBacks.Token
	if token == nil {
		return nil, fmt.Errorf("sprut token empty: %w", err)
	}

	urls, err := getUrls(respWithCallbacks.ApiCallBacks.Apis, sprutApiKey)
	if err != nil {
		return nil, fmt.Errorf("can't get urls: %w", err)
	}

	var sections []int64
	for i := sectionStart; i <= sectionEnd; i++ {
		sections = append(sections, i)
	}

	var stageForSprut *int64
	if stage != nil && *stage == 0 {
		stageForSprut = nil
	} else {
		stageForSprut = stage
	}

	dataForSprut := struct {
		OfficeID int64   `json:"office_id"`
		WhID     int64   `json:"wh_id"`
		Street   int64   `json:"street"`
		Stage    *int64  `json:"stage,omitempty"`
		Sections []int64 `json:"sections"`
	}{
		OfficeID: officeID,
		WhID:     whID,
		Street:   street,
		Stage:    stageForSprut,
		Sections: sections,
	}

	dataForSprutToSend, err := jsoniter.Marshal(dataForSprut)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for sprut call backs: %w", err)
	}

	sprutApiResponse, err := a.sprutClient.CallSprutAPIByUrls(ctx, *token, urls, dataForSprutToSend)
	if err != nil {
		return nil, fmt.Errorf("can't call sprut api: %w", err)
	}

	if sprutApiResponse == nil {
		return nil, nil
	}

	var sprutResponse models.SprutStoragePlacesResponse

	err = jsoniter.Unmarshal(sprutApiResponse, &sprutResponse)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal sprut response: %w", err)
	}

	var storagePlaces []models.StoragePlace
	for _, section := range sprutResponse.Data.Sections {
		for _, place := range section.Places {
			storagePlaces = append(storagePlaces, models.StoragePlace{
				PlaceID:   place.PlaceID,
				PlaceName: place.PlaceName,
			})
		}
	}

	return storagePlaces, nil
}

func (a *SprutAccessService) GetStreetsByStage(ctx *gin.Context, employeeID, officeID, whID, stage int64) ([]models.StreetsInfo, error) {
	const (
		sprutApiKey   = "sprutenv.office.streets.get_by_stage/v001"
		backendApiKey = "WhSprutenvAdministrationGetApi"
	)

	dataForGetCallBacks := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
	}{officeID, employeeID}

	dataForGetCallBacksToSend, err := jsoniter.Marshal(dataForGetCallBacks)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for get call backs: %w", err)
	}

	respWithCallbacks, err := a.backendApiClient.GetCallBacks(backendApiKey, dataForGetCallBacksToSend)
	if err != nil {
		return nil, fmt.Errorf("can't get callbacks: %w", err)
	}

	if respWithCallbacks == nil {
		return nil, fmt.Errorf("response with callback is empty: %w", err)
	}

	token := respWithCallbacks.ApiCallBacks.Token
	if token == nil {
		return nil, fmt.Errorf("sprut token empty: %w", err)
	}

	urls, err := getUrls(respWithCallbacks.ApiCallBacks.Apis, sprutApiKey)
	if err != nil {
		return nil, fmt.Errorf("can't get urls: %w", err)
	}

	dataForSprut := struct {
		OfficeID   int64 `json:"office_id"`
		EmployeeID int64 `json:"employee_id"`
		WhID       int64 `json:"wh_id"`
		Stage      int64 `json:"stage"`
	}{officeID, employeeID, whID, stage}

	dataForSprutToSend, err := jsoniter.Marshal(dataForSprut)
	if err != nil {
		return nil, fmt.Errorf("can't marshal data for sprut call backs: %w", err)
	}

	sprutApiResponse, err := a.sprutClient.CallSprutAPIByUrls(ctx, *token, urls, dataForSprutToSend)
	if err != nil {
		return nil, fmt.Errorf("can't call sprut api: %w", err)
	}

	if sprutApiResponse == nil {
		return nil, nil
	}

	var streetsResponse models.StreetsResponse

	err = jsoniter.Unmarshal(sprutApiResponse, &streetsResponse)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal sprut response: %w", err)
	}

	return streetsResponse.Data, nil
}

func getUrls(apis []models.Apis, apiKeySprut string) ([]string, error) {
	for i := range apis {
		if apis[i].Key == apiKeySprut {
			return apis[i].Urls, nil
		}
	}
	return nil, fmt.Errorf("urls empty")
}
