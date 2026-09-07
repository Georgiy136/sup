package access

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/models"
)

func TestDetermineAccessibleCategories(t *testing.T) {
	type args struct {
		externalActionsByTypeAction []models.ExternalActionWithCategory
		actionGroupsByTypeAction    []models.ActionGroupInfoWithCategory
		employeeExternalActions     []string
		employeeGroups              []int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "Проверка доступа через внешние экшены",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{
					{ExternalAction: "UploadingCategory.Manage", CategoryID: 13},
					{ExternalAction: "UploadingCategory.View", CategoryID: 14},
				},
				actionGroupsByTypeAction: []models.ActionGroupInfoWithCategory{},
				employeeExternalActions:  []string{"UploadingCategory.Manage"},
				employeeGroups:           []int64{},
			},
			want: []int64{13},
		},
		{
			name: "Проверка доступа через группы доступа",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{},
				actionGroupsByTypeAction: []models.ActionGroupInfoWithCategory{
					{GroupID: 1, CategoryID: 20},
					{GroupID: 2, CategoryID: 21},
				},
				employeeExternalActions: []string{},
				employeeGroups:          []int64{1},
			},
			want: []int64{20},
		},
		{
			name: "Проверка доступа через оба источника (внешние экшены и группы доступа)",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{
					{ExternalAction: "UploadingCategory.Manage", CategoryID: 10},
				},
				actionGroupsByTypeAction: []models.ActionGroupInfoWithCategory{
					{GroupID: 5, CategoryID: 12},
					{GroupID: 10, CategoryID: 17},
				},
				employeeExternalActions: []string{"UploadingCategory.Manage"},
				employeeGroups:          []int64{5},
			},
			want: []int64{10, 12},
		},
		{
			name: "Дублирующиеся категории должны быть уникальными",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{
					{ExternalAction: "UploadingCategory.Manage", CategoryID: 100},
					{ExternalAction: "UploadingCategory.View", CategoryID: 100},
				},
				actionGroupsByTypeAction: []models.ActionGroupInfoWithCategory{
					{GroupID: 1, CategoryID: 100},
					{GroupID: 10, CategoryID: 17},
				},
				employeeExternalActions: []string{"UploadingCategory.Manage", "UploadingCategory.View"},
				employeeGroups:          []int64{1},
			},
			want: []int64{100},
		},
		{
			name: "Нет доступа ни к одной категории",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{
					{ExternalAction: "UploadingCategory.Manage", CategoryID: 1},
				},
				actionGroupsByTypeAction: []models.ActionGroupInfoWithCategory{
					{GroupID: 99, CategoryID: 2},
				},
				employeeExternalActions: []string{"UploadingCategory.View"},
				employeeGroups:          []int64{88},
			},
			want: []int64{},
		},
		{
			name: "Пустые входные данные",
			args: args{
				externalActionsByTypeAction: []models.ExternalActionWithCategory{},
				actionGroupsByTypeAction:    []models.ActionGroupInfoWithCategory{},
				employeeExternalActions:     []string{},
				employeeGroups:              []int64{},
			},
			want: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewCategoryAccessResolver()
			got := r.DetermineAccessibleCategories(
				tt.args.externalActionsByTypeAction,
				tt.args.actionGroupsByTypeAction,
				tt.args.employeeExternalActions,
				tt.args.employeeGroups,
			)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
