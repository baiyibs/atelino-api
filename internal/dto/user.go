package dto

import (
	"atelino/internal/model"
	"time"
)

// UserIDRequest 用户 ID 请求
type UserIDRequest struct {
	// 用户 ID
	ID uint64 `uri:"id" binding:"required,min=1" example:"1"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	// 邮箱地址
	Email string `json:"email" binding:"required,email" example:"user@example.com" format:"email"`

	// 邮箱验证码
	Code string `json:"code" binding:"required" example:"123456"`

	// 用户名，长度 2-20
	Username string `json:"username" binding:"required,max=20" example:"John" minLength:"2" maxLength:"20"`

	// 密码，最少 8 位
	Password string `json:"password" binding:"required,min=8" example:"password123" minLength:"8"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	// 邮箱地址
	Email string `json:"email" binding:"required,email" example:"user@example.com" format:"email"`

	// 密码
	Password string `json:"password" binding:"required" example:"password123"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	// 页数，从 1 开始
	Page int `form:"page" binding:"omitempty,min=1" example:"1" default:"1" minimum:"1"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	// 刷新令牌
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	// 用户 ID
	UserID string `json:"-" binding:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	// 用户 ID
	ID uint64 `json:"id" example:"1"`

	// 邮箱地址
	Email string `json:"email" example:"user@example.com"`

	// 用户名
	Username string `json:"username" example:"John"`

	// 角色：user 或 admin
	Role string `json:"role" example:"user" enums:"user,admin"`

	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

func NewUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func NewUserResponses(users []model.User) []UserResponse {
	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, NewUserResponse(user))
	}
	return responses
}

// TokenResponse 令牌响应
type TokenResponse struct {
	// 访问令牌
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`

	// 刷新令牌
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}
