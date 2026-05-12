package dto

import (
	"atelino/internal/model"
	"time"
)

// CreateHitokotoRequest 创建一言请求
type CreateHitokotoRequest struct {
	// 一言内容
	Content string `json:"content" binding:"required" example:"千里之行，始于足下。"`
}

// HitokotoIDRequest 一言 ID 请求
type HitokotoIDRequest struct {
	// 一言 ID
	ID int `uri:"id" binding:"required,min=1" example:"1" minimum:"1"`
}

// HitokotoListRequest 一言列表请求
type HitokotoListRequest struct {
	// 页数，从 1 开始
	Page int `form:"page" binding:"omitempty,min=1" example:"1" default:"1" minimum:"1"`
}

// HitokotoResponse 一言响应
type HitokotoResponse struct {
	// 一言 ID
	ID uint64 `json:"id" example:"1"`

	// 一言内容
	Content string `json:"content" example:"千里之行，始于足下。"`

	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

func NewHitokotoResponse(hitokoto model.Hitokoto) HitokotoResponse {
	return HitokotoResponse{
		ID:        hitokoto.ID,
		Content:   hitokoto.Content,
		CreatedAt: hitokoto.CreatedAt,
	}
}

func NewHitokotoResponses(hitokotos []model.Hitokoto) []HitokotoResponse {
	responses := make([]HitokotoResponse, 0, len(hitokotos))
	for _, hitokoto := range hitokotos {
		responses = append(responses, NewHitokotoResponse(hitokoto))
	}
	return responses
}
