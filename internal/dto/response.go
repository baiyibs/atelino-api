package dto

// Response 通用响应体
type Response struct {
	// 状态码，200 表示成功
	Code int `json:"code" example:"200"`

	// 响应消息
	Message string `json:"message" example:"请求成功"`

	// 响应数据
	Data interface{} `json:"data"`
}
