package user

import (
	"backend/src/internal/auth"
	"backend/src/internal/database"
	"backend/src/internal/model"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidToken = errors.New("无效的刷新令牌")
	ErrTokenExpired = errors.New("刷新令牌已失效")
)

// RegisterTask 注册请求
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
	database.RedisClient.Del(ctx.Request.Context(), codeKey)

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

// LoginTask 登录请求, 返回 access_token 和 refresh_token
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
		log.Printf("用户登录时发生错误: %v", err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	const maxUserDevices = 3

	var accessToken string
	var rawRefresh string
	var refreshHash string

	err := database.GormDB.Transaction(func(tx *gorm.DB) error {
		var validTokens []model.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", user.ID, time.Now()).
			Order("created_at ASC").
			Find(&validTokens).Error; err != nil {
			return fmt.Errorf("查询有效令牌失败: %w", err)
		}

		// 限制用户令牌数量
		currentCount := len(validTokens)
		if currentCount >= maxUserDevices {
			revokeCount := currentCount - maxUserDevices + 1
			for i := 0; i < revokeCount && i < currentCount; i++ {
				validTokens[i].RevokedAt = new(time.Now())
				if err := tx.Save(&validTokens[i]).Error; err != nil {
					return fmt.Errorf("吊销旧令牌失败: %w", err)
				}
			}
		}

		var err error
		// 生成访问令牌
		accessToken, err = auth.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return fmt.Errorf("生成访问令牌失败: %w", err)
		}
		// 生成刷新令牌
		rawRefresh, refreshHash, err = auth.GenerateRefreshToken()
		if err != nil {
			return fmt.Errorf("生成刷新令牌失败: %w", err)
		}

		// 存储刷新令牌
		newToken := model.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshHash,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&newToken).Error; err != nil {
			return fmt.Errorf("存储新令牌失败: %w", err)
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "登录失败，请稍后重试。",
		})
		log.Printf("用户登录失败: %v", err)
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

// RefreshTask 使用 Refresh Token 刷新 Access Token
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

	var newAccessToken string
	var newRawRefresh string

	err := database.GormDB.Transaction(func(tx *gorm.DB) error {
		var oldToken model.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ?", hash).
			First(&oldToken).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTokenExpired
			}
			return err
		}

		if oldToken.RevokedAt != nil || time.Now().After(oldToken.ExpiresAt) {
			return ErrTokenExpired
		}

		var user model.User
		if err := tx.First(&user, oldToken.UserID).Error; err != nil {
			return fmt.Errorf("查询用户失败: %w", err)
		}

		accessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return fmt.Errorf("生成访问令牌失败: %w", err)
		}
		newAccessToken = accessToken

		rawRefresh, newHash, err := auth.GenerateRefreshToken()
		if err != nil {
			return fmt.Errorf("生成刷新令牌失败: %w", err)
		}
		newRawRefresh = rawRefresh

		oldToken.RevokedAt = new(time.Now())
		if err := tx.Save(&oldToken).Error; err != nil {
			return fmt.Errorf("吊销令牌失败: %w", err)
		}

		newToken := model.RefreshToken{
			UserID:    oldToken.UserID,
			TokenHash: newHash,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		if err := tx.Create(&newToken).Error; err != nil {
			return fmt.Errorf("存储新令牌失败: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			ctx.JSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: err.Error(),
			})
			return
		}
		if errors.Is(err, ErrTokenExpired) {
			ctx.JSON(http.StatusUnauthorized, model.Response{
				Code:    401,
				Message: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "服务器内部错误",
		})
		log.Printf("刷新令牌事务失败: %v", err)
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

// LogoutTask 登出, 吊销该用户的 Refresh Token
func LogoutTask(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	err := database.GormDB.Transaction(func(tx *gorm.DB) error {
		var tokens []model.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND revoked_at IS NULL", userID).
			Find(&tokens).Error; err != nil {
			return err
		}

		if len(tokens) > 0 {
			now := time.Now()
			for i := range tokens {
				tokens[i].RevokedAt = &now
			}
			if err := tx.Save(&tokens).Error; err != nil {
				return err
			}
		}

		return nil
	})

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
