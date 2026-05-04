package hitokoto

import (
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// 通过ID返回一条指定的一言
func GetHitokotoById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的一言id",
		})
		return
	}
	var hitokoto model.Hitokoto

	query := `SELECT id, content FROM hitokoto WHERE id = $1`
	if err := database.Pool.QueryRow(ctx.Request.Context(), query, id).Scan(&hitokoto.Id, &hitokoto.Content); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code:    404,
				Message: "没有找到对应的内容",
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

// 通过数据库返回一条随机一言
func GetHitokotoRandom(ctx *gin.Context) {
	var hitokoto model.Hitokoto
	query := `SELECT id, content FROM hitokoto ORDER BY RANDOM() LIMIT 1;`
	if err := database.Pool.QueryRow(ctx.Request.Context(), query).Scan(&hitokoto.Id, &hitokoto.Content); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code:    404,
				Message: "没有找到对应的内容",
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
