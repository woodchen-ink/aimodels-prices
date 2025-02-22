package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

// GetModelTypes 获取所有模型类型
func GetModelTypes(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query("SELECT type_key, type_label FROM model_type")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var types []models.ModelType
	for rows.Next() {
		var t models.ModelType
		if err := rows.Scan(&t.TypeKey, &t.TypeLabel); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		types = append(types, t)
	}

	c.JSON(200, types)
}

// CreateModelType 添加新的模型类型
func CreateModelType(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var newType models.ModelType
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(`
		INSERT INTO model_type (type_key, type_label)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE type_label = VALUES(type_label)
	`, newType.TypeKey, newType.TypeLabel)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, newType)
}
