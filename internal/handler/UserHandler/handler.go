package UserHandler

import (
	"backend/internal/database"
	"backend/internal/dto"
	"backend/internal/repository/UserRepository"
	"backend/internal/repository/ValidatorRepository"
	"backend/internal/service/UserService"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func newService() *UserService.Service {
	validatorRepo := ValidatorRepository.NewValidatorRepository(database.RedisClient)

	return UserService.New(
		UserRepository.NewUserRepository(database.GormDB),
		UserRepository.NewGormTransactionManager(database.GormDB),
		validatorRepo,
	)
}

func bindID(ctx *gin.Context) (dto.UserIDRequest, bool) {
	var request dto.UserIDRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return dto.UserIDRequest{}, false
	}
	return request, true
}

func bindPage(ctx *gin.Context) (dto.UserListRequest, bool) {
	var request dto.UserListRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return dto.UserListRequest{}, false
	}

	return request, true
}

func RegisterTask(ctx *gin.Context) {
	var request dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		log.Println(err)
		return
	}

	if err := newService().Register(request); err != nil {
		switch {
		case errors.Is(err, UserService.ErrVerificationCodeExpired):
			ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: err.Error()})
		case errors.Is(err, UserService.ErrInvalidVerificationCode):
			ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: err.Error()})
		case errors.Is(err, UserService.ErrEmailExists):
			ctx.JSON(http.StatusConflict, dto.Response{Code: 409, Message: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "注册失败"})
			log.Printf("注册失败: %v", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "注册成功"})
}

func LoginTask(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return
	}

	tokens, err := newService().Login(request)
	if err != nil {
		if errors.Is(err, UserService.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, dto.Response{Code: 401, Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "登录失败，请稍后重试。"})
		log.Printf("用户登录失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "登录成功", Data: tokens})
}

func GetUserByID(ctx *gin.Context) {
	request, ok := bindID(ctx)
	if !ok {
		return
	}

	user, err := newService().GetByID(request)
	if err != nil {
		if errors.Is(err, UserService.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, dto.Response{Code: 404, Message: "没有找到对应的用户"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "请求成功", Data: user})
}

func GetUserList(ctx *gin.Context) {
	request, ok := bindPage(ctx)
	if !ok {
		return
	}

	const pageSize = 20
	list, _, err := newService().List(request, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		log.Printf("获取用户列表失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "请求成功", Data: list})
}

func RefreshTask(ctx *gin.Context) {
	var request dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return
	}

	tokens, err := newService().Refresh(request)
	if err != nil {
		if errors.Is(err, UserService.ErrInvalidToken) || errors.Is(err, UserService.ErrTokenExpired) {
			ctx.JSON(http.StatusUnauthorized, dto.Response{Code: 401, Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "服务器内部错误"})
		log.Printf("刷新令牌失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "刷新成功", Data: tokens})
}

func LogoutTask(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Code: 401, Message: "未登录"})
		return
	}

	userIDString, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Code: 401, Message: "无效的登录状态"})
		return
	}

	if err := newService().Logout(dto.LogoutRequest{UserID: userIDString}); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "登出失败"})
		log.Printf("登出失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "登出成功"})
}
