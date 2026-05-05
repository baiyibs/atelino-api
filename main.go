package main

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/service/hitokoto"
	"backend/internal/service/user"
	"backend/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// 设置日志格式
	utils.SetupLog()

	// 初始化环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化JWT
	auth.InitJWT()

	// 初始化数据库
	if err := database.InitGorm(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseGorm()

	// 初始化路由
	router := gin.Default()

	// 不需要权限验证的接口
	GroupApi := router.Group("api")
	{
		GroupApi.POST("/login", user.Login)
		GroupApi.POST("/refresh", user.Refresh)

		GroupHitokoto := GroupApi.Group("hitokoto") // 一言
		{
			GroupHitokoto.POST("/", hitokoto.GetHitokotoRandom)
		}
	}

	// 需要权限验证的接口
	GroupAuth := router.Group("api")
	GroupAuth.Use(middleware.AuthMiddleware())
	{
		GroupApi.POST("/logout", user.Logout)
	}

	// 需要管理员权限才能访问的接口
	GroupAdmin := router.Group("api")
	GroupAdmin.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{
		GroupHitokoto := GroupAdmin.Group("hitokoto")
		{
			GroupHitokoto.POST("/getHitokotoList", hitokoto.GetHitokotoList)
			GroupHitokoto.POST("/getHitokotoById", hitokoto.GetHitokotoById)

			GroupHitokoto.POST("/insertHitokotoWithContent", hitokoto.InsertHitokotoWithContent)

			GroupHitokoto.POST("/deleteHitokotoById", hitokoto.DeleteHitokotoById)
		}
	}
	router.Run()
}
