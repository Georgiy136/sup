package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/local_errors"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

func (s *SupportFileManagerService) bindServiceError(ctx *gin.Context, err error, msg string) {
	if rd := local_errors.ToRestData(err); rd != nil {
		utils.BindRestData(ctx, *rd)
		return
	}
	s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("%s: %w", msg, err))
}
