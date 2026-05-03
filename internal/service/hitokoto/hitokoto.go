package hitokoto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 通过数据库返回一条随机一言
func Hitokoto(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "测试",
	})
}
