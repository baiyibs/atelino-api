package model

import "time"

// Hitokoto 一言
type Hitokoto struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Content   string    `gorm:"not null;uniqueIndex" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
