package model

import "time"

type User struct {
	ID                    string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Username              string     `gorm:"uniqueIndex;not null" json:"username"`
	Password              string     `gorm:"not null" json:"-"`
	Role                  string     `gorm:"default:user" json:"role"` // 权限组
	CreatedAt             time.Time  `json:"created_at"`
	RefreshTokenHash      string     `json:"-"` // Refresh Token (有效期一周)
	RefreshTokenExpiresAt *time.Time `json:"-"` // 过期时间
	RefreshTokenRevokedAt *time.Time `json:"-"` // 吊销时间
}
