package main

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/scheduler/cron"
	"backend/internal/scheduler/task"
	"backend/internal/service/hitokoto"
	"backend/internal/service/user"
	"backend/internal/service/verify"
	"backend/pkg/email"
	"backend/pkg/utils"
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
	email.InitSmtpService()

	// 初始化 Cron 调度器
	cron.InitScheduler()
	// 添加计划任务
	cron.Register("吊销刷新令牌").Daily(12, 0).Do(task.CleanupRevokedTokens)

	cron.StartScheduler()
	defer cron.StopScheduler()
	// 初始化路由
	router := gin.Default()

	// 不需要权限验证的接口
	GroupApi := router.Group("api")
	{
		GroupHitokoto := GroupApi.Group("hitokoto") // 一言
		{
			GroupHitokoto.GET("", hitokoto.GetHitokotoRandom)
		}
	}

	GroupVerify := router.Group("verify")
	{
		GroupVerify.POST("/send", verify.SendVerificationCode)
	}

	GroupAuth := router.Group("auth")
	{
		GroupAuth.POST("/login", user.LoginTask)
		GroupAuth.POST("/refresh", user.RefreshTask)
		GroupAuth.POST("/register", user.RegisterTask)
	}
	GroupAuthSecure := GroupAuth.Use(middleware.AuthMiddleware())
	{
		// 需要权限验证的接口
		GroupAuthSecure.POST("/logout", user.LogoutTask)
	}

	// 需要管理员权限才能访问的接口
	GroupAdmin := router.Group("api")
	GroupAdmin.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{
		GroupHitokoto := GroupAdmin.Group("hitokoto")
		{
			GroupHitokoto.GET("/list", hitokoto.GetHitokotoList)
			GroupHitokoto.GET("/:id", hitokoto.GetHitokotoById)

			GroupHitokoto.POST("", hitokoto.InsertHitokotoWithContent)

			GroupHitokoto.DELETE("/:id", hitokoto.DeleteHitokotoById)
		}
	}

	if err := router.Run(); err != nil {
		log.Fatalf("Gin 启动失败: %v", err)
	}
}
