package dto

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	// 邮箱地址
	To string `json:"to" binding:"required,email" example:"user@example.com" format:"email"`
}
