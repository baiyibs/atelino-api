package model

// 通用响应体
type Response struct {
	Code    int
	Message string
	Data    interface{}
}

// 一言
type Hitokoto struct {
	Id      int
	Content string
}
