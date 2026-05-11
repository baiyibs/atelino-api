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

// CreateHitokotoWithContent 根据请求体中的内容创建一条新的一言记录。
//
//	@Summary		添加一言
//	@Description	创建一条新的一言记录。如果内容已存在，则返回 409 冲突错误；其他数据库异常返回 500。
//	@Tags			一言
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateHitokotoRequest				true	"一言内容"
//	@Success		200		{object}	dto.Response{data=dto.HitokotoResponse}	"添加成功"
//	@Failure		400		{object}	dto.Response{}							"请求参数错误"
//	@Failure		409		{object}	dto.Response{}							"该一言已存在"
//	@Failure		500		{object}	dto.Response{}							"数据库错误"
//	@Router			/hitokoto [post]
func CreateHitokotoWithContent(ctx *gin.Context) {
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

// DeleteHitokotoById 根据 ID 删除一条一言记录。
//
//	@Summary		删除一言
//	@Description	传入一言的 ID，从数据库中删除对应的记录。若 ID 不存在则返回 404，数据库异常则返回 500。
//	@Tags			一言
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int				true	"一言 ID"
//	@Success		200	{object}	dto.Response{}	"删除成功"
//	@Failure		400	{object}	dto.Response{}	"请求参数错误"
//	@Failure		401	{object}	dto.Response{}	"未授权"
//	@Failure		404	{object}	dto.Response{}	"没有找到对应的一言"
//	@Failure		500	{object}	dto.Response{}	"数据库错误"
//	@Router			/hitokoto/{id} [delete]
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

// GetHitokotoList 获取所有一言
//
//	@Summary		获取一言列表
//	@Description	从数据库中查询所有的一言记录，数据库异常则返回 500。
//	@Tags			一言
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int											false	"页数"
//	@Success		200		{object}	dto.Response{data=[]dto.HitokotoResponse}	"请求成功"
//	@Failure		401		{object}	dto.Response{}								"未授权"
//	@Failure		500		{object}	dto.Response{}								"数据库错误"
//	@Router			/hitokoto/list [get]
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

// GetHitokotoById 根据 ID 获取一条一言记录。
//
//	@Summary		获取一言
//	@Description	传入一言的 ID，从数据库中查询对应的记录。若 ID 不存在则返回 404，数据库异常则返回 500。
//	@Tags			一言
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"一言 ID"
//	@Success		200	{object}	dto.Response{data=dto.HitokotoResponse}	"请求成功"
//	@Failure		400	{object}	dto.Response{}							"请求参数错误"
//	@Failure		401	{object}	dto.Response{}							"未授权"
//	@Failure		404	{object}	dto.Response{}							"没有找到对应的一言"
//	@Failure		500	{object}	dto.Response{}							"数据库错误"
//	@Router			/hitokoto/{id} [get]
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

// GetHitokotoRandom 随机获取一条一言记录。
//
//	@Summary		随机获取一言
//	@Description	从数据库中随机获取一条一言记录，数据库异常则返回 500。
//	@Tags			一言
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.Response{data=dto.HitokotoResponse}	"请求成功"
//	@Failure		404	{object}	dto.Response{}							"没有找到对应的一言"
//	@Failure		500	{object}	dto.Response{}							"数据库错误"
//	@Router			/hitokoto [get]
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
