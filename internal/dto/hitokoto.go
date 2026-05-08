package dto

import (
	"backend/internal/model"
	"time"
)

type CreateHitokotoRequest struct {
	Content string `json:"content" binding:"required"`
}

type HitokotoIDRequest struct {
	ID int `uri:"id" binding:"required,min=1"`
}

type HitokotoListRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`
}

type HitokotoResponse struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
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
