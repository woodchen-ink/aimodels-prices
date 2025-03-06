package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// GetProviders 获取所有模型厂商
func GetProviders(c *gin.Context) {
	cacheKey := "providers"

	// 尝试从缓存获取
	if cachedData, found := database.GlobalCache.Get(cacheKey); found {
		if providers, ok := cachedData.([]models.Provider); ok {
			c.JSON(http.StatusOK, providers)
			return
		}
	}

	var providers []models.Provider

	if err := database.DB.Order("id").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	// 存入缓存，有效期30分钟
	database.GlobalCache.Set(cacheKey, providers, 30*time.Minute)

	c.JSON(http.StatusOK, providers)
}

// CreateProvider 创建模型厂商
func CreateProvider(c *gin.Context) {
	var provider models.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查ID是否已存在
	var existingProvider models.Provider
	result := database.DB.Where("id = ?", provider.ID).First(&existingProvider)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID already exists"})
		return
	}

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	currentUser := user.(*models.User)

	// 设置创建者
	provider.CreatedBy = currentUser.Username

	// 创建记录
	if err := database.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	// 清除缓存
	database.GlobalCache.Delete("providers")

	c.JSON(http.StatusCreated, provider)
}

// UpdateProvider 更新模型厂商
func UpdateProvider(c *gin.Context) {
	oldID := c.Param("id")
	var provider models.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找现有记录
	var existingProvider models.Provider
	if err := database.DB.Where("id = ?", oldID).First(&existingProvider).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// 如果ID发生变化，需要同时更新price表中的引用
	if oldID != strconv.FormatUint(uint64(provider.ID), 10) {
		// 开始事务
		tx := database.DB.Begin()

		// 更新price表中的channel_type
		if err := tx.Model(&models.Price{}).Where("channel_type = ?", oldID).Update("channel_type", provider.ID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price references"})
			return
		}

		// 更新price表中的temp_channel_type
		if err := tx.Model(&models.Price{}).Where("temp_channel_type = ?", oldID).Update("temp_channel_type", provider.ID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price temp references"})
			return
		}

		// 删除旧记录
		if err := tx.Delete(&existingProvider).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old provider"})
			return
		}

		// 创建新记录
		provider.CreatedAt = time.Now()
		provider.UpdatedAt = time.Now()
		if err := tx.Create(&provider).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new provider"})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}
	} else {
		// 如果ID没有变化，直接更新
		existingProvider.Name = provider.Name
		existingProvider.Icon = provider.Icon
		if err := database.DB.Save(&existingProvider).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
			return
		}
		provider = existingProvider
	}

	// 清除缓存
	database.GlobalCache.Delete("providers")

	c.JSON(http.StatusOK, provider)
}

// UpdateProviderStatus 更新模型厂商状态
func UpdateProviderStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required,oneof=approved rejected"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()

	if input.Status == "approved" {
		// 如果是批准，将临时字段的值更新到正式字段
		result := database.DB.Exec(`
			UPDATE provider 
			SET name = COALESCE(temp_name, name),
				icon = COALESCE(temp_icon, icon),
				status = ?,
				updated_at = ?,
				temp_name = NULL,
				temp_icon = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider status"})
			return
		}
	} else {
		// 如果是拒绝，清除临时字段
		result := database.DB.Exec(`
			UPDATE provider 
			SET status = ?,
				updated_at = ?,
				temp_name = NULL,
				temp_icon = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Status updated successfully",
		"status":     input.Status,
		"updated_at": now,
	})
}

// DeleteProvider 删除模型厂商
func DeleteProvider(c *gin.Context) {
	id := c.Param("id")

	// 查找现有记录
	var provider models.Provider
	if err := database.DB.Where("id = ?", id).First(&provider).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// 检查是否有价格记录使用此厂商
	var count int64
	if err := database.DB.Model(&models.Price{}).Where("channel_type = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check provider usage"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete provider that is in use"})
		return
	}

	// 删除记录
	if err := database.DB.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	// 清除缓存
	database.GlobalCache.Delete("providers")

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}
