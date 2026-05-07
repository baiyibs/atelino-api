package hitokoto

import (
	"backend/src/internal/database"
	"backend/src/internal/model"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type hitokotoRequest struct {
	ID int `json:"id" binding:"required"`
}

// InsertHitokotoWithContent 插入新的一言
func InsertHitokotoWithContent(ctx *gin.Context) {
	var request struct {
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: fmt.Sprintf("请求错误: %s", err.Error()),
		})
		return
	}

	if len(request.Content) == 0 {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "需要添加的一言不能为空",
		})
		return
	}

	hitokoto := model.Hitokoto{
		Content: request.Content,
	}
	if err := database.GormDB.Create(&hitokoto).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.JSON(http.StatusConflict, model.Response{Code: 409, Message: "该一言已存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "添加成功",
		Data:    hitokoto,
	})
}

// DeleteHitokotoById 通过ID删除一条一言
func DeleteHitokotoById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	result := database.GormDB.Where("id = ?", id).Delete(&model.Hitokoto{})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, model.Response{
			Code:    404,
			Message: "没有找到对应的一言",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "删除成功",
	})
}

// GetHitokotoList 获取一言列表
func GetHitokotoList(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	const pageSize = 20
	offset := (page - 1) * pageSize

	var list []model.Hitokoto
	var total int64

	if err := database.GormDB.Model(&model.Hitokoto{}).Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		log.Printf("获取一言列表总数失败: %v", err)
		return
	}

	if err := database.GormDB.Order("id asc").Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		log.Printf("获取一言列表失败: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "请求成功",
		Data:    list,
	})
}

// GetHitokotoById 通过ID返回一条指定的一言
func GetHitokotoById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	var hitokoto model.Hitokoto
	if err := database.GormDB.Where("id = ?", id).First(&hitokoto).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code:    404,
				Message: "没有找到对应的一言",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "请求成功",
		Data:    hitokoto,
	})
}

// GetHitokotoRandom 通过数据库返回一条随机一言
func GetHitokotoRandom(ctx *gin.Context) {
	var hitokoto model.Hitokoto

	if err := database.GormDB.Order("RANDOM()").Limit(1).First(&hitokoto).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code:    404,
				Message: "没有找到对应的一言",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "请求成功",
		Data:    hitokoto,
	})
}
