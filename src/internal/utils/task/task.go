package task

import (
	"backend/src/internal/database"
	"backend/src/internal/model"
	"log"
	"time"
)

// CleanupRevokedTokens 删除已经被吊销一周的刷新令牌
func CleanupRevokedTokens() {
	cutoff := time.Now().AddDate(0, 0, -7)
	result := database.GormDB.Where("revoked_at IS NOT NULL AND revoked_at < ?", cutoff).
		Delete(&model.RefreshToken{})

	if result.Error != nil {
		log.Printf("清理已经过期一周的刷新令牌失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("已清理 %d 条吊销超过一周的刷新令牌记录", result.RowsAffected)
	}
}
