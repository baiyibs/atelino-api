package main

import (
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/service/hitokoto"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// 初始化环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 初始化路由
	router := gin.Default()

	GroupApi := router.Group("api")
	{
		GroupHitokoto := GroupApi.Group("hitokoto") // 一言
		{
			GroupHitokoto.GET("/", hitokoto.GetHitokotoRandom)
		}
	}
	GroupAdmin := router.Group("api")
	GroupAdmin.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{
		GroupHitokoto := GroupAdmin.Group("hitokoto") // 一言
		{
			GroupHitokoto.GET("/list", hitokoto.GetHitokotoList)
			GroupHitokoto.GET("/:id", hitokoto.GetHitokotoById)

			GroupHitokoto.POST("/", hitokoto.InsertHitokotoWithContent)

			GroupHitokoto.DELETE("/:id", hitokoto.DeleteHitokotoById)
		}
	}
	router.Run()
}
