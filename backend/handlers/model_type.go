package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// GetModelTypes 获取所有模型类型
func GetModelTypes(c *gin.Context) {
	var types []models.ModelType

	// 使用GORM查询所有模型类型，按排序字段和键值排序
	if err := database.DB.Order("sort_order ASC, type_key ASC").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types)
}

// CreateModelType 添加新的模型类型
func CreateModelType(c *gin.Context) {
	var newType models.ModelType
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用GORM创建新记录
	if err := database.DB.Create(&newType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newType)
}

// UpdateModelType 更新模型类型
func UpdateModelType(c *gin.Context) {
	typeKey := c.Param("key")
	var updateType models.ModelType
	if err := c.ShouldBindJSON(&updateType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找现有记录
	var existingType models.ModelType
	if err := database.DB.Where("type_key = ?", typeKey).First(&existingType).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model type not found"})
		return
	}

	// 如果key发生变化，需要删除旧记录并创建新记录
	if typeKey != updateType.TypeKey {
		// 开始事务
		tx := database.DB.Begin()

		// 删除旧记录
		if err := tx.Delete(&existingType).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old model type"})
			return
		}

		// 创建新记录
		if err := tx.Create(&updateType).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new model type"})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}
	} else {
		// 直接更新
		existingType.TypeLabel = updateType.TypeLabel
		existingType.SortOrder = updateType.SortOrder
		if err := database.DB.Save(&existingType).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model type"})
			return
		}
		updateType = existingType
	}

	c.JSON(http.StatusOK, updateType)
}

// DeleteModelType 删除模型类型
func DeleteModelType(c *gin.Context) {
	typeKey := c.Param("key")

	// 查找现有记录
	var existingType models.ModelType
	if err := database.DB.Where("type_key = ?", typeKey).First(&existingType).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model type not found"})
		return
	}

	// 检查是否有价格记录使用此类型
	var count int64
	if err := database.DB.Model(&models.Price{}).Where("model_type = ?", typeKey).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check model type usage"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete model type that is in use"})
		return
	}

	// 删除记录
	if err := database.DB.Delete(&existingType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model type deleted successfully"})
}
