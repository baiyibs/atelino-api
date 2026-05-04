package model

// 通用响应体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 一言
type Hitokoto struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}
