package app

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/router"
	"backend/internal/scheduler/cron"
	"backend/internal/scheduler/task"
	"backend/pkg/email"
	"backend/pkg/utils"
	"log"

	"github.com/joho/godotenv"
)

func Run() {

	// 设置日志格式
	utils.SetupLog()

	// 加载环境变量
	loadEnv()

	// 初始化依赖
	initDependencies()
	defer shutdownDependencies()

	// 注册定时任务
	registerJobs()

	// 注册路由
	if err := router.New().Run(); err != nil {
		log.Fatalf("Gin 启动失败: %v", err)
	}
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
		log.Println("没有发现 .env 文件，使用环境变量")
	}
}

func initDependencies() {
	auth.InitJWT()

	if err := database.InitGorm(); err != nil {
		log.Fatalf("Gorm 初始化失败: %v", err)
	}

	if err := database.InitRedis(); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}

	email.InitSmtpService()

	cron.InitScheduler()
	cron.StartScheduler()
}

func shutdownDependencies() {
	cron.StopScheduler()
	database.CloseRedis()
	database.CloseGorm()
}

func registerJobs() {
	cron.Register("清理无效令牌").
		Daily(12, 0).
		Do(task.CleanupInvalidRefreshTokens)
}
