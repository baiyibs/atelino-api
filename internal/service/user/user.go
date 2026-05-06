package user

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 注册请求
func RegisterTask(ctx *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
		Username string `json:"username" binding:"required,max=20"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		log.Println(err)
		return
	}

	codeKey := "verify:code:" + request.Email
	code, err := database.RedisClient.Get(ctx.Request.Context(), codeKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ctx.JSON(http.StatusBadRequest, model.Response{
				Code:    400,
				Message: "验证码无效或已经过期，请重新获取。",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	if code != request.Code {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "验证码无效",
		})
		return
	}

	// 查询用户是否存在
	var existUser model.User
	if err := database.GormDB.Where("email = ?", request.Email).First(&existUser).Error; err == nil {
		ctx.JSON(http.StatusConflict, model.Response{
			Code:    409,
			Message: "该邮箱已注册",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "系统内部错误",
		})
		log.Printf("加密密码失败: %v", err)
		return
	}

	// 删除验证码
	database.RedisClient.Del(ctx.Request.Context(), codeKey).Err()

	// 创建用户
	user := model.User{
		Email:    request.Email,
		Username: request.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	// 存储用户
	if err := database.GormDB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "注册失败",
		})
		log.Printf("创建用户失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "注册成功",
	})
}

// 登录请求, 返回 access_token 和 refresh_token
func LoginTask(ctx *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	var user model.User

	if err := database.GormDB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// 使用 RefreshTask Token 刷新 Access Token
func RefreshTask(ctx *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	hash := auth.HashRefreshToken(request.RefreshToken)

	// 开启事务
	tx := database.GormDB.Begin()
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}
	// 事务异常时回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("refresh_token_hash = ?", hash).
		First(&user).Error; err != nil {
		tx.Rollback()
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
		tx.Rollback()
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "刷新令牌已失效",
		})
		return
	}

	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成访问令牌失败",
		})
		return
	}

	newRawRefresh, newRefreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "生成刷新令牌失败",
		})
		return
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	updateData := map[string]interface{}{
		"refresh_token_hash":       newRefreshHash,
		"refresh_token_expires_at": newExpiresAt,
		"refresh_token_revoked_at": nil,
	}

	if err := tx.Model(&user).Updates(updateData).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "更新刷新令牌失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
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

// 登出, 吊销该用户的 Refresh Token
func LogoutTask(ctx *gin.Context) {
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
