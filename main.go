package main

import (
	"backend/internal/database"
	"backend/internal/dev"
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

	// 生成测试用户
	dev.GenerateTestTokens()

	// 初始化路由
	router := gin.Default()

	GroupApi := router.Group("api")
	{
		GroupHitokoto := GroupApi.Group("hitokoto") // 一言
		{
			GroupHitokoto.POST("/", hitokoto.GetHitokotoRandom)
		}
	}
	GroupAdmin := router.Group("api")
	GroupAdmin.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{
		GroupHitokoto := GroupAdmin.Group("hitokoto") // 一言
		{
			GroupHitokoto.POST("/getHitokotoList", hitokoto.GetHitokotoList)
			GroupHitokoto.POST("/getHitokotoById", hitokoto.GetHitokotoById)

			GroupHitokoto.POST("/insertHitokotoWithContent", hitokoto.InsertHitokotoWithContent)

			GroupHitokoto.POST("/deleteHitokotoById", hitokoto.DeleteHitokotoById)
		}
	}
	router.Run()
}
