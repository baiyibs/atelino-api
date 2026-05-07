package verify

import (
	"backend/internal/database"
	"backend/internal/model"
	"backend/pkg/email"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SendVerificationCode(ctx *gin.Context) {
	var request struct {
		To string `json:"to" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	// 限制发送频率
	cooldownKey := "verify:cooldown:" + request.To
	ok, err := database.RedisClient.SetNX(ctx.Request.Context(), cooldownKey, 1, time.Minute).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusTooManyRequests, model.Response{
			Code:    429,
			Message: "请求过于频繁",
		})
		return
	}

	// 生成验证码
	code, err := email.GenerateVerificationCode()
	if err != nil {
		database.RedisClient.Del(ctx.Request.Context(), cooldownKey)
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成验证码失败",
		})
		log.Printf("生成验证码时发生错误: %v", err)
		return
	}

	// 存储验证码
	codeKey := "verify:code:" + request.To
	if err := database.RedisClient.Set(ctx.Request.Context(), codeKey, code, 5*time.Minute).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	// 发送验证码
	if err := email.SendVerificationCode(request.To, code); err != nil {
		database.RedisClient.Del(ctx.Request.Context(), cooldownKey)
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "邮件发送失败，请稍后再试。",
		})
		log.Printf("发送邮件失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "发送成功",
	})
}
