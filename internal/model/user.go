package model

import "time"

type User struct {
	ID                    string     `json:"id"`
	Username              string     `json:"username"`
	Password              string     `json:"-"`
	Role                  string     `json:"role"` // 权限组
	CreatedAt             string     `json:"created_at"`
	RefreshTokenHash      string     `json:"-"` // Refresh Token (有效期一周)
	RefreshTokenExpiresAt *time.Time `json:"-"` // 过期时间
	RefreshTokenRevokedAt *time.Time `json:"-"` // 吊销时间
}
