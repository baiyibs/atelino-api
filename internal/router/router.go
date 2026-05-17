package router

import (
	"atelino/internal/handler/HitokotoHandler"
	"atelino/internal/handler/UserHandler"
	"atelino/internal/handler/ValidatorHandler"
	"atelino/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipQueryString: true,
	}))
	// 测试用
	r.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: false,
			},
		),
	)

	api := r.Group("api")
	{
		userGroup := api.Group("user")
		{
			userGroup.GET("/:id", UserHandler.GetUserByID)
		}
		hitokotoGroup := api.Group("hitokoto")
		{
			hitokotoGroup.GET("", HitokotoHandler.GetHitokotoRandom)
		}
	}

	verifyGroup := api.Group("verify")
	{
		verifyGroup.POST("/send", ValidatorHandler.SendVerificationCode)
	}

	authGroup := api.Group("auth")
	{
		authGroup.POST("/login", UserHandler.LoginTask)
		authGroup.POST("/refresh", UserHandler.RefreshTask)
		authGroup.POST("/register", UserHandler.RegisterTask)
	}

	authSecureGroup := authGroup.Use(middleware.AuthMiddleware())
	{
		authSecureGroup.POST("/logout", UserHandler.LogoutTask)
	}

	adminGroup := r.Group("api")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminRequired())
	{
		userGroup := adminGroup.Group("/user")
		{
			userGroup.GET("/list", UserHandler.GetUserList)
		}

		hitokotoGroup := adminGroup.Group("hitokoto")
		{
			hitokotoGroup.GET("/list", HitokotoHandler.GetHitokotoList)
			hitokotoGroup.GET("/:id", HitokotoHandler.GetHitokotoById)
			hitokotoGroup.POST("", HitokotoHandler.CreateHitokotoWithContent)
			hitokotoGroup.DELETE("/:id", HitokotoHandler.DeleteHitokotoById)
		}
	}

	return r
}
