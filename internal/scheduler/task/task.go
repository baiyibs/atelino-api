package task

import (
	"atelino/internal/database"
	"atelino/internal/model"
	"log"
	"time"
)

// CleanupInvalidRefreshTokens  清理所有需要删除的刷新令牌
//   - 过期超过一周
//   - 吊销超过一周
func CleanupInvalidRefreshTokens() {
	cutoff := time.Now().UTC().AddDate(0, 0, -7)

	result := database.GormDB.Where("revoked_at IS NOT NULL AND revoked_at < ?", cutoff).
		Or("expires_at IS NOT NULL AND expires_at < ?", cutoff).
		Delete(&model.RefreshToken{})

	if result.Error != nil {
		log.Printf("清理无效的刷新令牌失败: %v", result.Error)
		return
	}
	log.Printf("已清理 %d 条无效的的刷新令牌记录", result.RowsAffected)
}
