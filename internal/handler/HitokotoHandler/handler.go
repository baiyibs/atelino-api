package HitokotoHandler

import (
	"atelino/internal/database"
	"atelino/internal/dto"
	"atelino/internal/repository/HitokotoRepository"
	"atelino/internal/service/HitokotoService"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func newService() *HitokotoService.Service {
	return HitokotoService.New(HitokotoRepository.NewHitokotoRepository(database.GormDB))
}

func bindID(ctx *gin.Context) (dto.HitokotoIDRequest, bool) {
	var request dto.HitokotoIDRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return dto.HitokotoIDRequest{}, false
	}
	return request, true
}

func bindPage(ctx *gin.Context) (dto.HitokotoListRequest, bool) {
	var request dto.HitokotoListRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Code: 400, Message: "无效的请求"})
		return dto.HitokotoListRequest{}, false
	}
	return request, true
}

// InsertHitokotoWithContent godoc
//
//	@Summary		添加一言
//	@Description	创建一条新的一言记录，如果内容已存在则返回冲突错误
//	@Tags			一言管理
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateHitokotoRequest					true	"一言内容"
//	@Success		200		{object}	dto.Response{data=dto.HitokotoIDRequest}	"添加成功"
//	@Failure		400		{object}	dto.Response								"请求参数错误"
//	@Failure		409		{object}	dto.Response								"该一言已存在"
//	@Failure		500		{object}	dto.Response								"数据库错误"
//	@Router			/hitokoto [post]
func InsertHitokotoWithContent(ctx *gin.Context) {
	var request dto.CreateHitokotoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    400,
			Message: fmt.Sprintf("请求错误: %s", err.Error()),
		})
		return
	}

	hitokoto, err := newService().Create(request)
	if err != nil {
		if errors.Is(err, HitokotoService.ErrDuplicate) {
			ctx.JSON(http.StatusConflict, dto.Response{Code: 409, Message: "该一言已存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "添加成功", Data: hitokoto})
}

func DeleteHitokotoById(ctx *gin.Context) {
	request, ok := bindID(ctx)
	if !ok {
		return
	}

	if err := newService().DeleteByID(request); err != nil {
		if errors.Is(err, HitokotoService.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, dto.Response{Code: 404, Message: "没有找到对应的一言"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "删除成功"})
}

func GetHitokotoList(ctx *gin.Context) {
	request, ok := bindPage(ctx)
	if !ok {
		return
	}

	const pageSize = 20
	list, _, err := newService().List(request, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		log.Printf("获取一言列表失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "请求成功", Data: list})
}

func GetHitokotoById(ctx *gin.Context) {
	request, ok := bindID(ctx)
	if !ok {
		return
	}

	hitokoto, err := newService().GetByID(request)
	if err != nil {
		if errors.Is(err, HitokotoService.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, dto.Response{Code: 404, Message: "没有找到对应的一言"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "请求成功", Data: hitokoto})
}

func GetHitokotoRandom(ctx *gin.Context) {
	hitokoto, err := newService().Random()
	if err != nil {
		if errors.Is(err, HitokotoService.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, dto.Response{Code: 404, Message: "没有找到对应的一言"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Code: 500, Message: "数据库错误"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Code: 200, Message: "请求成功", Data: hitokoto})
}
