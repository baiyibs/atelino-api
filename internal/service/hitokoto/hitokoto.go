package hitokoto

import (
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type hitokotoRequest struct {
	ID int `json:"id" binding:"required"`
}

// 插入新的一言
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

	sql := `INSERT INTO hitokoto (content) VALUES ($1) RETURNING id`
	var newID int
	if err := database.Pool.QueryRow(ctx.Request.Context(), sql, request.Content).Scan(&newID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 唯一冲突码
			ctx.JSON(http.StatusConflict, model.Response{
				Code:    409,
				Message: "该一言已存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	hitokoto := model.Hitokoto{
		Id:      newID,
		Content: request.Content,
	}
	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "添加成功",
		Data:    hitokoto,
	})
}

// 通过ID删除一条一言
func DeleteHitokotoById(ctx *gin.Context) {
	var request hitokotoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	sql := `DELETE FROM hitokoto WHERE id = $1`
	cmdTag, err := database.Pool.Exec(ctx.Request.Context(), sql, request.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, model.Response{
			Code:    404,
			Message: "没有找到要删除的一言",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "删除成功",
	})
}

// 获取一言列表
func GetHitokotoList(ctx *gin.Context) {
	sql := `SELECT id, content FROM hitokoto ORDER BY id ASC`
	rows, err := database.Pool.Query(ctx.Request.Context(), sql)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}
	defer rows.Close()

	var list []model.Hitokoto
	for rows.Next() {
		var row model.Hitokoto
		if err := rows.Scan(&row.Id, &row.Content); err != nil {
			ctx.JSON(http.StatusInternalServerError, model.Response{
				Code:    500,
				Message: "数据库错误",
			})
			return
		}
		list = append(list, row)
	}
	ctx.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "请求成功",
		Data:    list,
	})
}

// 通过ID返回一条指定的一言
func GetHitokotoById(ctx *gin.Context) {
	var request hitokotoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "无效的请求",
		})
		return
	}

	var hitokoto model.Hitokoto
	sql := `SELECT id, content FROM hitokoto WHERE id = $1`
	if err := database.Pool.QueryRow(ctx.Request.Context(), sql, request.ID).Scan(&hitokoto.Id, &hitokoto.Content); err != nil {
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
	sql := `SELECT id, content FROM hitokoto ORDER BY RANDOM() LIMIT 1;`
	if err := database.Pool.QueryRow(ctx.Request.Context(), sql).Scan(&hitokoto.Id, &hitokoto.Content); err != nil {
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
