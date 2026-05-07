package model

import "time"

// Hitokoto 一言
type Hitokoto struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Content   string    `gorm:"not null;uniqueIndex" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
