package dto

type SendVerificationCodeRequest struct {
	To string `json:"to" binding:"required,email"`
}
