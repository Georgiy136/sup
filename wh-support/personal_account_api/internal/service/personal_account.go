package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	client "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/personal_account_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/personal_account_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/personal_account_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PersonalAccount struct {
	hrEmployeePhotoClient *client.HrEmployeePhotoApiClient
	employeeInfoClient    *client.EmployeeInfoApiClient
	errBuilder            core_errors.ErrorBuilder
}

func NewPersonalAccount() *PersonalAccount {
	return &PersonalAccount{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil)}
}

func (p *PersonalAccount) Configure(ctx context.Context, configs configs.Config) {
	p.hrEmployeePhotoClient = client.NewHrEmployeePhotoApiClient(ctx, configs)
	p.employeeInfoClient = client.NewEmployeeInfoApiClient(ctx, configs)
}

func (p *PersonalAccount) GetPersonalAccountInfoV2(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	info, err := p.employeeInfoClient.GetEmployeeSelfFullName(employeeID)
	if err != nil {
		logrus.Warnf("can't get employee info: %v", err)
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't get employee info: %w", err))
		return
	}

	if info == nil {
		logrus.Warnf("nil from employee_info: %v", common.ErrEmployeeUnknown)
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("can't find employee: %w", common.ErrEmployeeUnknown))
		return
	}

	splittedName := strings.Split(info.EmployeeFullName, " ")
	if len(splittedName) < 1 {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), errors.New("invalid employee name parameter"))
		return
	}

	photo, err := p.hrEmployeePhotoClient.GetEmployeePhoto(employeeID)
	if err != nil {
		logrus.Warnf("can't get employee photo: %v", err)
	}

	data := &models.PersonalInfoRespV2{
		EmployeeId: employeeID,
		Photo:      photo,
	}

	if len(splittedName) >= 3 {
		data.Patronymic = strings.Join(splittedName[2:], " ")
	}
	if len(splittedName) > 1 {
		data.Name = splittedName[1]
	}
	data.Surname = splittedName[0]

	utils.BindObjectToRestData(ctx, data)
}

func (p *PersonalAccount) GetGroupsByEmployee(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		p.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.DatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("hr.employeegroups_getbyemployee")

	rd := repository.GetRestDataFromDb(pg)
	utils.BindRestData(ctx, rd)
}
