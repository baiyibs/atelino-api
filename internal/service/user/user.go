package user

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// 登录请求, 返回 access_token 和 refresh_token
func Login(ctx *gin.Context) {
	var request loginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	var user model.User

	if err := database.GormDB.Where("username = ?", request.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code:    404,
				Message: "用户名或密码错误",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	// 生成访问令牌
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成访问令牌失败",
		})
		return
	}

	// 生成刷新令牌
	rawRefresh, refreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成刷新令牌失败",
		})
		return
	}

	// 存储刷新令牌
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	user.RefreshTokenHash = refreshHash
	user.RefreshTokenExpiresAt = &expiresAt
	user.RefreshTokenRevokedAt = nil

	if err := database.GormDB.Model(&user).
		Select("RefreshTokenHash", "RefreshTokenExpiresAt", "RefreshTokenRevokedAt").Updates(user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "更新刷新令牌失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "登录成功",
		Data: gin.H{
			"access_token":  accessToken,
			"refresh_token": rawRefresh,
		},
	})
}

// 使用 Refresh Token 刷新 Access Token
func Refresh(ctx *gin.Context) {
	var request refreshRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	hash := auth.HashRefreshToken(request.RefreshToken)

	var user model.User
	if err := database.GormDB.Where("refresh_token_hash = ?", hash).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: "无效的刷新令牌",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	if user.RefreshTokenRevokedAt != nil || user.RefreshTokenExpiresAt == nil || time.Now().After(*user.RefreshTokenExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "刷新令牌已失效",
		})
		return
	}

	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成访问令牌失败",
		})
		return
	}

	newRawRefresh, newRefreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成刷新令牌失败",
		})
		return
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	user.RefreshTokenHash = newRefreshHash
	user.RefreshTokenExpiresAt = &newExpiresAt
	user.RefreshTokenRevokedAt = nil

	if err := database.GormDB.Model(&user).
		Select("RefreshTokenHash", "RefreshTokenExpiresAt", "RefreshTokenRevokedAt").Updates(user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "更新刷新令牌失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "刷新成功",
		Data: gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRawRefresh,
		},
	})
}

// 吊销 Refresh Token
func Logout(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	var user model.User
	if err := database.GormDB.Where("id = ?", userID).First(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	now := time.Now()
	user.RefreshTokenRevokedAt = &now
	if err := database.GormDB.Model(&user).Select("RefreshTokenRevokedAt").Updates(user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "登出失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "登出成功",
	})
}
