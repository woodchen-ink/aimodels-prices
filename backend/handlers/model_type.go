package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

// GetModelTypes 获取所有模型类型
func GetModelTypes(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query("SELECT type_key, type_label, sort_order FROM model_type ORDER BY sort_order ASC, type_key ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var types []models.ModelType
	for rows.Next() {
		var t models.ModelType
		if err := rows.Scan(&t.TypeKey, &t.TypeLabel, &t.SortOrder); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		types = append(types, t)
	}

	c.JSON(http.StatusOK, types)
}

// CreateModelType 添加新的模型类型
func CreateModelType(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var newType models.ModelType
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(`
		INSERT INTO model_type (type_key, type_label, sort_order)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE type_label = VALUES(type_label), sort_order = VALUES(sort_order)
	`, newType.TypeKey, newType.TypeLabel, newType.SortOrder)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newType)
}

// UpdateModelType 更新模型类型
func UpdateModelType(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	typeKey := c.Param("key")

	var updateType models.ModelType
	if err := c.ShouldBindJSON(&updateType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果key发生变化，需要删除旧记录并创建新记录
	if typeKey != updateType.TypeKey {
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		// 删除旧记录
		_, err = tx.Exec("DELETE FROM model_type WHERE type_key = ?", typeKey)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old model type"})
			return
		}

		// 创建新记录
		_, err = tx.Exec(`
			INSERT INTO model_type (type_key, type_label, sort_order)
			VALUES (?, ?, ?)
		`, updateType.TypeKey, updateType.TypeLabel, updateType.SortOrder)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new model type"})
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}
	} else {
		// 直接更新
		_, err := db.Exec(`
			UPDATE model_type 
			SET type_label = ?, sort_order = ?
			WHERE type_key = ?
		`, updateType.TypeLabel, updateType.SortOrder, typeKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model type"})
			return
		}
	}

	c.JSON(http.StatusOK, updateType)
}

// DeleteModelType 删除模型类型
func DeleteModelType(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	typeKey := c.Param("key")

	// 检查是否有价格记录使用此类型
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM price WHERE model_type = ?", typeKey).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check model type usage"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete model type that is in use"})
		return
	}

	_, err = db.Exec("DELETE FROM model_type WHERE type_key = ?", typeKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model type deleted successfully"})
}
