package main

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/service/hitokoto"
	"backend/internal/service/user"
	"backend/internal/service/verify"
	"backend/internal/utils"
	"backend/internal/utils/email"
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
		log.Println(err)
		log.Println("没有找到.env文件,将使用环境变量")
	}

	// 初始化 JWT
	auth.InitJWT()

	// 初始化 Gorm
	if err := database.InitGorm(); err != nil {
		log.Fatalf("Gorm 初始化失败: %v", err)
	}
	defer database.CloseGorm()

	// 初始化 Redis
	if err := database.InitRedis(); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}
	defer database.CloseRedis()

	// 初始化邮箱配置
	email.InitStmpService()

	// 初始化路由
	router := gin.Default()

	// 不需要权限验证的接口
	GroupApi := router.Group("api")
	{
		GroupApi.POST("/login", user.LoginTask)
		GroupApi.POST("/refresh", user.RefreshTask)
		GroupApi.POST("/register", user.RegisterTask)

		GroupVerify := GroupApi.Group("verify") // 验证
		{
			GroupVerify.POST("/SendVerificationCode", verify.SendVerificationCode)
		}

		GroupHitokoto := GroupApi.Group("hitokoto") // 一言
		{
			GroupHitokoto.POST("/", hitokoto.GetHitokotoRandom)
		}
	}

	// 需要权限验证的接口
	GroupAuth := router.Group("api")
	GroupAuth.Use(middleware.AuthMiddleware())
	{
		GroupAuth.POST("/logout", user.LogoutTask)
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
