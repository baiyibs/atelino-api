package model

import "time"

type User struct {
	ID                    uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email                 string     `gorm:"uniqueIndex;not null" json:"email"`
	Username              string     `gorm:"not null;size:20" json:"username"`
	Password              string     `gorm:"not null" json:"-"`
	Role                  string     `gorm:"default:user" json:"role"` // 权限组 (user 和 admin)
	CreatedAt             time.Time  `json:"created_at"`
	RefreshTokenHash      string     `json:"-"` // Refresh Token (有效期一周)
	RefreshTokenExpiresAt *time.Time `json:"-"` // 过期时间
	RefreshTokenRevokedAt *time.Time `json:"-"` // 吊销时间
}
