package hitokoto

import (
	"backend/internal/database"
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的id",
		})
		return
	}
	var content string

	query := `SELECT content FROM hitokoto WHERE id = $1`
	if err := database.Pool.QueryRow(ctx.Request.Context(), query, id).Scan(&content); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "没有找到对应的内容",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": content,
	})
}

// 通过数据库返回一条随机一言
func GetHitokotoRandom(ctx *gin.Context) {
	var content string
	query := `SELECT content FROM hitokoto ORDER BY RANDOM() LIMIT 1;`
	if err := database.Pool.QueryRow(ctx.Request.Context(), query).Scan(&content); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "没有找到对应的内容",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": content,
	})
}
