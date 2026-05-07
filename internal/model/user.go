package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"` // UserID
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Username  string    `gorm:"not null;size:20;index" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:user" json:"role"` // 权限组 (user 和 admin)
	CreatedAt time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	UserID    uint       `gorm:"not null;index:idx_user_revoke_exp,priority:1"` // 关联用户
	TokenHash string     `gorm:"not null;uniqueIndex"`                          // 令牌哈希，用于查找
	ExpiresAt time.Time  `gorm:"index:idx_user_revoke_exp,priority:3"`          // 过期时间
	RevokedAt *time.Time `gorm:"index:idx_user_revoke_exp,priority:2"`          // 吊销时间（NULL 表示有效）
	CreatedAt time.Time
}
