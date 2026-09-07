package test

import (
	"reflect"
	"testing"

	createoperationmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_operations/models"
	createstorageplacetypemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_type/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
)

func TestDecodeMapToStructureWithoutErrorUnset(t *testing.T) {
	rateLimit := 12.5

	// задание столбцов таблицы
	testcases := []struct {
		testName   string
		input      map[string]any
		expect     any
		parseModel any
		expectErr  bool
	}{
		{
			testName:  "Test decode",
			expectErr: false,
			input: map[string]any{
				"prodtype_code":              "code1",
				"prodtype_name":              "name1",
				"prodtype_part_id":           map[string]any{"id": int64(123)},
				"is_use_improver_multiplier": true,
				"rate_limit":                 rateLimit,
			},
			expect: createoperationmodels.ExtTicketInfoCreationOperation{
				ProdtypeCode: "code1",
				ProdtypeName: "name1",
				ProdtypePartID: struct {
					ID int64 `mapstructure:"id"`
				}{ID: 123},
				IsUseImproverMultiplier: true,
				RateLimit:               &rateLimit,
			},
			parseModel: createoperationmodels.ExtTicketInfoCreationOperation{},
		},
		{
			testName:  "Test decode",
			expectErr: false,
			input: map[string]any{
				"prodtype_code":              "code1",
				"prodtype_name":              "name1",
				"prodtype_part_id":           map[string]any{"id": int64(123)},
				"is_use_improver_multiplier": true,
				"rate_limit":                 nil,
			},
			expect: createoperationmodels.ExtTicketInfoCreationOperation{
				ProdtypeCode: "code1",
				ProdtypeName: "name1",
				ProdtypePartID: struct {
					ID int64 `mapstructure:"id"`
				}{ID: 123},
				IsUseImproverMultiplier: true,
				RateLimit:               nil,
			},
			parseModel: createoperationmodels.ExtTicketInfoCreationOperation{},
		},
		{
			testName:  "Test decode with expect error",
			expectErr: true,
			input: map[string]any{
				"prodtype_code":              "code1",
				"prodtype_name":              "name1",
				"prodtype_part_id":           map[string]any{"id": int64(123)},
				"is_use_improver_multiplier": true,
				"rate_limit":                 &rateLimit, // задано значение
			},
			expect: createoperationmodels.ExtTicketInfoCreationOperation{
				ProdtypeCode: "code1",
				ProdtypeName: "name1",
				ProdtypePartID: struct {
					ID int64 `mapstructure:"id"`
				}{ID: 123},
				IsUseImproverMultiplier: true,
				RateLimit:               nil, // значение не задано
			},
			parseModel: createoperationmodels.ExtTicketInfoCreationOperation{},
		},
		{
			testName:  "Test decode without pointer field",
			expectErr: false,
			input: map[string]any{
				"prodtype_code":              "code2",
				"prodtype_name":              "name1",
				"prodtype_part_id":           map[string]any{"id": int64(123)},
				"is_use_improver_multiplier": true,
			},
			expect: createoperationmodels.ExtTicketInfoCreationOperation{
				ProdtypeCode: "code2",
				ProdtypeName: "name1",
				ProdtypePartID: struct {
					ID int64 `mapstructure:"id"`
				}{ID: 123},
				IsUseImproverMultiplier: true,
				RateLimit:               nil, // значение задано
			},
			parseModel: createoperationmodels.ExtTicketInfoCreationOperation{},
		},
		{
			testName:  "Test decode without field with expected error",
			expectErr: true,
			input: map[string]any{
				//"prodtype_code":              "code2",
				"prodtype_name":              "name1",
				"prodtype_part_id":           map[string]any{"id": int64(123)},
				"is_use_improver_multiplier": true,
			},
			expect: createoperationmodels.ExtTicketInfoCreationOperation{
				ProdtypeCode: "code2",
				ProdtypeName: "name1",
				ProdtypePartID: struct {
					ID int64 `mapstructure:"id"`
				}{ID: 123},
				IsUseImproverMultiplier: true,
				RateLimit:               nil,
			},
			parseModel: createoperationmodels.ExtTicketInfoCreationOperation{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.testName, func(t *testing.T) {
			if err := handlerutils.DecodeMapToStructureWithoutErrorUnset(tt.input, &tt.parseModel); err != nil {
				t.Fatalf("can't decode map to structure without err unset: %v", err)
			}

			if !reflect.DeepEqual(tt.parseModel, tt.expect) && !tt.expectErr {
				t.Fatalf("got %v; want %v", tt.parseModel, tt.expect)
			}
		})
	}
}

func TestDecodeMapToStructureWithErrorUnset(t *testing.T) {
	// задание столбцов таблицы
	testcases := []struct {
		testName   string
		input      map[string]any
		expect     any
		parseModel any
		expectErr  bool
	}{
		{
			testName:  "Test decode without field with expected error",
			expectErr: false,
			input: map[string]any{
				"place_type_name": "test",
				"sticker_type":    "test",
				"sticker_prefix":  map[string]any{"id": "$Pb"},
			},
			expect: createstorageplacetypemodels.ExtTicketInfoForCreateStoragePlaceType{
				PlaceTypeName: "test",
				StickerType:   "test",
				StickerPrefix: struct {
					ID string `mapstructure:"id"`
				}{ID: "$Pb"},
			},
			parseModel: createstorageplacetypemodels.ExtTicketInfoForCreateStoragePlaceType{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.testName, func(t *testing.T) {
			if err := handlerutils.DecodeMapToStructureWithErrorUnset(tt.input, &tt.parseModel); err != nil {
				t.Fatalf("can't decode map to structure with err unset: %v", err)
			}

			if !reflect.DeepEqual(tt.parseModel, tt.expect) && !tt.expectErr {
				t.Fatalf("got %v; want %v", tt.parseModel, tt.expect)
			}
		})
	}
}
