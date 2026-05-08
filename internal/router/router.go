package router

import (
	"backend/internal/handler/HitokotoHandler"
	"backend/internal/handler/UserHandler"
	"backend/internal/handler/ValidatorHandler"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	api := r.Group("api")
	{
		hitokotoGroup := api.Group("hitokoto")
		{
			hitokotoGroup.GET("", HitokotoHandler.GetHitokotoRandom)
		}
	}

	verifyGroup := r.Group("verify")
	{
		verifyGroup.POST("/send", ValidatorHandler.SendVerificationCode)
	}

	authGroup := r.Group("auth")
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
		hitokotoGroup := adminGroup.Group("hitokoto")
		{
			hitokotoGroup.GET("/list", HitokotoHandler.GetHitokotoList)
			hitokotoGroup.GET("/:id", HitokotoHandler.GetHitokotoById)
			hitokotoGroup.POST("", HitokotoHandler.InsertHitokotoWithContent)
			hitokotoGroup.DELETE("/:id", HitokotoHandler.DeleteHitokotoById)
		}
	}

	return r
}
