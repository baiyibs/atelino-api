package ValidatorHandler

import (
	"atelino/internal/database"
	"atelino/internal/dto"
	"atelino/internal/repository/ValidatorRepository"
	"atelino/internal/service/ValidatorService"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func newService() *ValidatorService.Service {
	return ValidatorService.New(ValidatorRepository.NewValidatorRepository(database.RedisClient))
}

// SendVerificationCode 发送验证码
//
//	@Summary		发送验证码
//	@Description	向指定的邮箱发送验证码用于注册或其他用途。
//	@Tags			验证
//	@Accept			json
//	@Produce		json
//	@ID				sendVerificationCode
//	@Param			request	body		dto.SendVerificationCodeRequest	true	"邮箱地址"
//	@Success		200		{object}	dto.Response{}					"发送成功"
//	@Failure		400		{object}	dto.Response{}					"请求参数错误"
//	@Failure		429		{object}	dto.Response{}					"请求过于频繁，请稍后再试"
//	@Failure		500		{object}	dto.Response{}					"发送验证码失败"
//	@Router			/verify/send [post]
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
