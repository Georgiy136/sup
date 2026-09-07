package access

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
)

func TestDetermineCategoryStatusAccess(t *testing.T) {
	var (
		statusID1 = "1"
		statusID2 = "2"
	)

	type args struct {
		employeeGroups          []int64
		employeeExternalActions []string
		accessDataList          []*models.AccessData
	}
	tests := []struct {
		name string
		args args
		want []models.CategoryStatusAccess
	}{
		{
			name: "Проверка доступа через внешние экшены",
			args: args{
				employeeGroups: []int64{},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
							{ExternalAction: "action2", CategoryID: 200, StatusID: &statusID2},
							{ExternalAction: "action1", CategoryID: 200, StatusID: nil},
						},
						ActionGroups: []models.ActionGroupWithCategory{},
					},
				},
				employeeExternalActions: []string{"action1"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
				{CategoryID: 200, StatusID: nil},
			},
		},
		{
			name: "Проверка доступа через группы доступа",
			args: args{
				employeeGroups: []int64{1},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{},
						ActionGroups: []models.ActionGroupWithCategory{
							{GroupID: 1, CategoryID: 100, StatusID: &statusID1},
							{GroupID: 1, CategoryID: 100, StatusID: nil},
							{GroupID: 2, CategoryID: 200, StatusID: &statusID2},
						},
					},
				},
				employeeExternalActions: []string{},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
				{CategoryID: 100, StatusID: nil},
			},
		},
		{
			name: "Проверка доступа через внешние экшены и группы доступа",
			args: args{
				employeeGroups: []int64{1},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 200, StatusID: &statusID2},
						},
						ActionGroups: []models.ActionGroupWithCategory{
							{GroupID: 1, CategoryID: 100, StatusID: &statusID1},
						},
					},
				},
				employeeExternalActions: []string{"action1"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
				{CategoryID: 200, StatusID: &statusID2},
			},
		},
		{
			name: "Дублирующиеся категории со статусами должны быть уникальными",
			args: args{
				employeeGroups: []int64{1},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
							{ExternalAction: "action2", CategoryID: 100, StatusID: &statusID1},
						},
						ActionGroups: []models.ActionGroupWithCategory{
							{GroupID: 1, CategoryID: 100, StatusID: &statusID1},
						},
					},
				},
				employeeExternalActions: []string{"action1", "action2"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
			},
		},
		{
			name: "Нет доступа ни к одной категории",
			args: args{
				employeeGroups: []int64{},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
						},
						ActionGroups: []models.ActionGroupWithCategory{
							{GroupID: 1, CategoryID: 200, StatusID: &statusID2},
						},
					},
				},
				employeeExternalActions: []string{},
			},
			want: []models.CategoryStatusAccess{},
		},
		{
			name: "Пустой массив AccessData",
			args: args{
				employeeGroups:          []int64{},
				employeeExternalActions: []string{"action1"},
				accessDataList:          []*models.AccessData{{}},
			},
			want: []models.CategoryStatusAccess{},
		},
		{
			name: "Пустые списки доступов",
			args: args{
				employeeGroups: []int64{},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{},
						ActionGroups:    []models.ActionGroupWithCategory{},
					},
				},
				employeeExternalActions: []string{},
			},
			want: []models.CategoryStatusAccess{},
		},
		{
			name: "Передан массив AccessData с дубликатами",
			args: args{
				employeeGroups: []int64{},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
						},
						ActionGroups: []models.ActionGroupWithCategory{},
					},
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
							{ExternalAction: "action1", CategoryID: 200, StatusID: &statusID2},
						},
						ActionGroups: []models.ActionGroupWithCategory{},
					},
				},
				employeeExternalActions: []string{"action1"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
				{CategoryID: 200, StatusID: &statusID2},
			},
		},
		{
			name: "Массив AccessData с доступами через группы и внешние экшены",
			args: args{
				employeeGroups: []int64{1},
				accessDataList: []*models.AccessData{
					{
						ExternalActions: []models.ExternalActionWithCategory{},
						ActionGroups: []models.ActionGroupWithCategory{
							{GroupID: 1, CategoryID: 100, StatusID: &statusID1},
						},
					},
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 200, StatusID: &statusID2},
						},
						ActionGroups: []models.ActionGroupWithCategory{},
					},
				},
				employeeExternalActions: []string{"action1"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
				{CategoryID: 200, StatusID: &statusID2},
			},
		},
		{
			name: "Внутри AccessData nil элемент",
			args: args{
				employeeGroups: []int64{},
				accessDataList: []*models.AccessData{
					nil,
					{
						ExternalActions: []models.ExternalActionWithCategory{
							{ExternalAction: "action1", CategoryID: 100, StatusID: &statusID1},
						},
						ActionGroups: []models.ActionGroupWithCategory{},
					},
				},
				employeeExternalActions: []string{"action1"},
			},
			want: []models.CategoryStatusAccess{
				{CategoryID: 100, StatusID: &statusID1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineCategoryStatusAccess(tt.args.employeeGroups, tt.args.employeeExternalActions, tt.args.accessDataList...)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCollectAccessibleCategoriesByGroups(t *testing.T) {
	tests := []struct {
		name           string
		actionGroups   []models.ActionGroupWithCategory
		employeeGroups []int64
		want           map[int64]struct{}
	}{
		{
			name: "simple match",
			actionGroups: []models.ActionGroupWithCategory{
				{GroupID: 1, CategoryID: 10},
				{GroupID: 2, CategoryID: 20},
			},
			employeeGroups: []int64{1},
			want:           map[int64]struct{}{10: {}},
		},
		{
			name: "no match",
			actionGroups: []models.ActionGroupWithCategory{
				{GroupID: 3, CategoryID: 30},
			},
			employeeGroups: []int64{1},
			want:           map[int64]struct{}{},
		},
		{
			name: "multiple matches",
			actionGroups: []models.ActionGroupWithCategory{
				{GroupID: 1, CategoryID: 10},
				{GroupID: 2, CategoryID: 20},
			},
			employeeGroups: []int64{1, 2},
			want:           map[int64]struct{}{10: {}, 20: {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(map[int64]struct{})
			CollectAccessibleCategoriesByGroups(got, tt.actionGroups, tt.employeeGroups)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectAccessibleCategoriesByExternalActions(t *testing.T) {
	tests := []struct {
		name                    string
		externalActions         []models.ExternalActionWithCategory
		employeeExternalActions []string
		want                    map[int64]struct{}
	}{
		{
			name: "simple match",
			externalActions: []models.ExternalActionWithCategory{
				{ExternalAction: "act1", CategoryID: 10},
				{ExternalAction: "act2", CategoryID: 20},
			},
			employeeExternalActions: []string{"act1"},
			want:                    map[int64]struct{}{10: {}},
		},
		{
			name: "no match",
			externalActions: []models.ExternalActionWithCategory{
				{ExternalAction: "act3", CategoryID: 30},
			},
			employeeExternalActions: []string{"act1"},
			want:                    map[int64]struct{}{},
		},
		{
			name: "multiple matches",
			externalActions: []models.ExternalActionWithCategory{
				{ExternalAction: "act1", CategoryID: 10},
				{ExternalAction: "act2", CategoryID: 20},
			},
			employeeExternalActions: []string{"act1", "act2"},
			want:                    map[int64]struct{}{10: {}, 20: {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(map[int64]struct{})
			CollectAccessibleCategoriesByExternalActions(got, tt.externalActions, tt.employeeExternalActions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAccessByPolicy(t *testing.T) {
	tests := []struct {
		name                    string
		accessPolicy            models.ResourceAccessPolicy
		employeeGroups          []int64
		employeeExternalActions []string
		expected                bool
	}{
		{
			name: "access by group",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{1, 2},
				ExternalActions: []string{"approve"},
			},
			employeeGroups:          []int64{2},
			employeeExternalActions: []string{},
			expected:                true,
		},
		{
			name: "access by external action",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{1},
				ExternalActions: []string{"approve", "reject"},
			},
			employeeGroups:          []int64{},
			employeeExternalActions: []string{"reject"},
			expected:                true,
		},
		{
			name: "access by group and external action",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{1},
				ExternalActions: []string{"approve"},
			},
			employeeGroups:          []int64{1},
			employeeExternalActions: []string{"approve"},
			expected:                true,
		},
		{
			name: "no matching groups or external actions",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{1},
				ExternalActions: []string{"approve"},
			},
			employeeGroups:          []int64{2},
			employeeExternalActions: []string{"reject"},
			expected:                false,
		},
		{
			name: "empty employee groups and external actions",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{1},
				ExternalActions: []string{"approve"},
			},
			employeeGroups:          []int64{},
			employeeExternalActions: []string{},
			expected:                false,
		},
		{
			name: "empty access policy",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{},
				ExternalActions: []string{},
			},
			employeeGroups:          []int64{1},
			employeeExternalActions: []string{"approve"},
			expected:                false,
		},
		{
			name: "partial match only group list present",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{3},
				ExternalActions: []string{},
			},
			employeeGroups:          []int64{3},
			employeeExternalActions: []string{},
			expected:                true,
		},
		{
			name: "partial match only external action list present",
			accessPolicy: models.ResourceAccessPolicy{
				AccessGroups:    []int64{},
				ExternalActions: []string{"view"},
			},
			employeeGroups:          []int64{},
			employeeExternalActions: []string{"view"},
			expected:                true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasAccessByPolicy(
				tt.accessPolicy,
				tt.employeeGroups,
				tt.employeeExternalActions,
			)

			if result != tt.expected {
				t.Errorf(
					"unexpected result: got %v, want %v",
					result,
					tt.expected,
				)
			}
		})
	}
}
