package ValidatorHandler

import (
	"backend/internal/database"
	"backend/internal/dto"
	"backend/internal/repository/ValidatorRepository"
	"backend/internal/service/ValidatorService"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func newService() *ValidatorService.Service {
	return ValidatorService.New(ValidatorRepository.NewValidatorRepository(database.RedisClient))
}

func SendVerificationCode(ctx *gin.Context) {
	var request dto.SendVerificationCodeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return
	}

	if err := newService().SendCode(request); err != nil {
		if errors.Is(err, ValidatorService.ErrCooldown) {
			ctx.JSON(http.StatusTooManyRequests, dto.Response{Code: 429, Message: "请求过于频繁"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "发送验证码失败"})
		log.Printf("发送验证码失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "发送成功"})
}
