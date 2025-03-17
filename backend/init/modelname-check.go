package init

import (
	"aimodels-prices/database"
	"aimodels-prices/models"
	"log"
)

// CheckDuplicateModelNames 检查数据库中是否存在重复的模型名称，如果有则保留最新的
func CheckDuplicateModelNames() error {
	log.Println("开始检查重复的模型名称...")
	db := database.DB
	if db == nil {
		return nil
	}

	// 查找所有具有重复模型名称的厂商ID和模型名称组合
	var duplicates []struct {
		ChannelType uint   `json:"channel_type"`
		Model       string `json:"model"`
		Count       int    `json:"count"`
	}

	if err := db.Raw(`
		SELECT channel_type, model, COUNT(*) as count
		FROM price
		GROUP BY channel_type, model
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error; err != nil {
		return err
	}

	if len(duplicates) == 0 {
		log.Println("没有发现重复的模型名称")
		return nil
	}

	log.Printf("发现 %d 组重复的模型名称，正在处理...", len(duplicates))
	processedCount := 0

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 处理每一组重复
	for _, dup := range duplicates {
		// 查找具有相同厂商ID和模型名称的所有记录
		var prices []models.Price
		if err := tx.Where("channel_type = ? AND model = ?", dup.ChannelType, dup.Model).Order("updated_at DESC").Find(&prices).Error; err != nil {
			tx.Rollback()
			return err
		}

		if len(prices) <= 1 {
			continue // 安全检查，实际上这不应该发生
		}

		// 保留最新的记录（按更新时间排序后的第一个），删除其他记录
		latestID := prices[0].ID
		log.Printf("保留最新的记录: ID=%v, 模型=%s, 厂商ID=%d, 更新时间=%v",
			latestID, dup.Model, dup.ChannelType, prices[0].UpdatedAt)

		// 收集要删除的ID
		var idsToDelete []uint
		for i := 1; i < len(prices); i++ {
			idsToDelete = append(idsToDelete, prices[i].ID)
			log.Printf("删除重复记录: ID=%v, 模型=%s, 厂商ID=%d, 更新时间=%v",
				prices[i].ID, dup.Model, dup.ChannelType, prices[i].UpdatedAt)
		}

		// 删除重复记录
		if len(idsToDelete) > 0 {
			if err := tx.Delete(&models.Price{}, idsToDelete).Error; err != nil {
				tx.Rollback()
				return err
			}
			processedCount += len(idsToDelete)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清除缓存
	database.GlobalCache.Clear()

	log.Printf("重复模型名称处理完成，共删除 %d 条重复记录", processedCount)
	return nil
}
