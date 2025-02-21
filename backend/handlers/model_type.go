package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// ModelType 模型类型结构
type ModelType struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// GetModelTypes 获取所有模型类型
func GetModelTypes(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query("SELECT key, label FROM model_type")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var types []ModelType
	for rows.Next() {
		var t ModelType
		if err := rows.Scan(&t.Key, &t.Label); err != nil {
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

	var newType ModelType
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(`
		INSERT INTO model_type (key, label)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET label = excluded.label
	`, newType.Key, newType.Label)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, newType)
}
