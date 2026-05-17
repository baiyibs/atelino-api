package dto

// Response 通用响应体
type Response struct {
	// 状态码，200 表示成功
	Code int `json:"code" example:"200"`

	// 响应消息
	Message string `json:"message"`

	// 响应数据
	Data interface{} `json:"data"`
}

// PaginatedResponse 分页列表响应体
type PaginatedResponse struct {
	// 列表数据
	List interface{} `json:"list"`

	// 总记录数
	Total int64 `json:"total"`
}
