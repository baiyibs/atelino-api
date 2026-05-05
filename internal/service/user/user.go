package user

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
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
	sql := `SELECT id, username, password, role FROM users WHERE username = $1`
	err := database.Pool.QueryRow(ctx.Request.Context(), sql, request.Username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, model.Response{
				Code:    401,
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
	updateSQL := `UPDATE users SET refresh_token_hash = $1, refresh_token_expires_at = $2, refresh_token_revoked_at = NULL WHERE id = $3`
	_, err = database.Pool.Exec(ctx.Request.Context(), updateSQL, refreshHash, expiresAt, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
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

	var userID, role string
	var expiresAt time.Time
	var revokedAt *time.Time

	query := `SELECT id, role, refresh_token_expires_at, refresh_token_revoked_at FROM users WHERE refresh_token_hash = $1`
	if err := database.Pool.QueryRow(ctx.Request.Context(), query, hash).Scan(&userID, &role, &expiresAt, &revokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的刷新令牌"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库错误"})
		return
	}

	if revokedAt != nil || time.Now().After(expiresAt) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌已失效"})
		return
	}

	newAccessToken, err := auth.GenerateAccessToken(userID, role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成访问令牌失败"})
		return
	}

	newRawRefresh, newRefreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成刷新令牌失败"})
		return
	}
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	updateSQL := `UPDATE users SET refresh_token_hash = $1, refresh_token_expires_at = $2, refresh_token_revoked_at = NULL WHERE id = $3`
	_, err = database.Pool.Exec(ctx.Request.Context(), updateSQL, newRefreshHash, newExpiresAt, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新刷新令牌失败"})
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
	var request refreshRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	hash := auth.HashRefreshToken(request.RefreshToken)
	updateSQL := `UPDATE users SET refresh_token_revoked_at = NOW() WHERE refresh_token_hash = $1`
	_, err := database.Pool.Exec(ctx.Request.Context(), updateSQL, hash)
	if err != nil {
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
