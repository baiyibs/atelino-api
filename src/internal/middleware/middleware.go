package middleware

import (
	"backend/src/internal/auth"
	"backend/src/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 验证JWT，提取用户信息
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: "你没有权限访问该请求!",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: "令牌格式错误",
			})
			return
		}

		claims, err := auth.ValidateAccessToken(parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: "无效或已经过期的令牌",
			})
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists || role != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, model.Response{
				Code:    403,
				Message: "你没有权限访问该请求!",
			})
			return
		}
		ctx.Next()
	}
}
