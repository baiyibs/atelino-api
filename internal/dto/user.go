package dto

import (
	"atelino/internal/model"
	"time"
)

type UserIDRequest struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	Username string `json:"username" binding:"required,max=20"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserListRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	UserID string `json:"-" binding:"required"`
}

type UserResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
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

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
