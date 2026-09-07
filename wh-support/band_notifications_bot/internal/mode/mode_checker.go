package mode

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

type employeeInfoRepo interface {
	GetAllowedEmployeesByGroup(groupID int64) ([]models.GroupDataResponse, error)
}

type ModeChecker struct {
	employeeInfoRepo employeeInfoRepo
	testMode         bool
	testGroups       []int64
}

func NewModeChecker(employeeInfoRepo employeeInfoRepo) *ModeChecker {
	return &ModeChecker{
		employeeInfoRepo: employeeInfoRepo,
	}
}

func (m *ModeChecker) Configure(ctx context.Context, cfg configs.Config) {
	exist, raw := cfg.GetByServiceKey("mode_config")
	var config struct {
		TestMode   bool    `json:"test_mode"`
		TestGroups []int64 `json:"test_groups"`
	}
	if !exist {
		logrus.Infof("mode_config not exists, use production mode")
		return
	}

	if err := jsoniter.Unmarshal(raw, &config); err != nil {
		logrus.Panicf("cannot unmarshal mode_config: %v", err)
	}
	m.testMode = config.TestMode
	m.testGroups = config.TestGroups

	logrus.Debugf("ModeChecker configured: TestMode=%t, loaded test employees %+v", m.testMode, m.GetTestEmployees())
}

func (m *ModeChecker) IsTestMode() bool {
	return m.testMode
}

func (m *ModeChecker) GetTestEmployees() []int64 {
	allowedTestEmployees := make(map[int64]struct{})
	for i := range m.testGroups {
		dbResp, err := m.employeeInfoRepo.GetAllowedEmployeesByGroup(m.testGroups[i])
		if err != nil {
			logrus.Errorf("cannot get allowed employees for group: %d, error: %v", m.testGroups[i], err)
			continue
		}
		if len(dbResp) == 0 {
			logrus.Errorf("no allowed employees in group %d", m.testGroups[i])
			continue
		}
		employeesInGroup := dbResp[0]

		for _, employee := range employeesInGroup.EmployeeInfo {
			allowedTestEmployees[employee.EmployeeID] = struct{}{}
		}
	}
	return maps.Keys(allowedTestEmployees)
}

func (m *ModeChecker) IsAllowedInTestMode(employeeID int64, employeesList []int64) bool {
	for i := range employeesList {
		if employeeID == employeesList[i] {
			return true
		}
	}
	return false
}
