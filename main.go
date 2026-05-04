package main

import (
	"backend/internal/database"
	"backend/internal/service/hitokoto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 定义路由
	router := gin.Default()
	api := router.Group("api")
	api.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	api.GET("/hitokoto/:id", hitokoto.GetHitokotoById)
	api.GET("/hitokoto", hitokoto.GetHitokotoRandom)
	router.Run()
}
