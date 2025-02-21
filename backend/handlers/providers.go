package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

// GetProviders 获取所有模型厂商
func GetProviders(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query(`
		SELECT id, name, icon, created_at, updated_at, created_by
		FROM provider ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}
	defer rows.Close()

	var providers []models.Provider
	for rows.Next() {
		var provider models.Provider
		if err := rows.Scan(
			&provider.ID, &provider.Name, &provider.Icon,
			&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan provider"})
			return
		}
		providers = append(providers, provider)
	}

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
	db := c.MustGet("db").(*sql.DB)
	var existingID int
	err := db.QueryRow("SELECT id FROM provider WHERE id = ?", provider.ID).Scan(&existingID)
	if err != sql.ErrNoRows {
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

	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO provider (id, name, icon, created_at, updated_at, created_by) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		provider.ID, provider.Name, provider.Icon, now, now, currentUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	provider.CreatedAt = now
	provider.UpdatedAt = now
	provider.CreatedBy = currentUser.Username

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

	db := c.MustGet("db").(*sql.DB)

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// 如果ID发生变化，需要同时更新price表中的引用
	if oldID != strconv.FormatUint(uint64(provider.ID), 10) {
		// 更新price表中的channel_type
		_, err = tx.Exec("UPDATE price SET channel_type = ? WHERE channel_type = ?", provider.ID, oldID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price references"})
			return
		}

		// 更新price表中的temp_channel_type
		_, err = tx.Exec("UPDATE price SET temp_channel_type = ? WHERE temp_channel_type = ?", provider.ID, oldID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price temp references"})
			return
		}

		// 删除旧记录
		_, err = tx.Exec("DELETE FROM provider WHERE id = ?", oldID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old provider"})
			return
		}

		// 插入新记录
		_, err = tx.Exec(`
			INSERT INTO provider (id, name, icon, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, provider.ID, provider.Name, provider.Icon)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new provider"})
			return
		}
	} else {
		// 如果ID没有变化，直接更新
		_, err = tx.Exec(`
			UPDATE provider 
			SET name = ?, icon = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, provider.Name, provider.Icon, oldID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

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

	db := c.MustGet("db").(*sql.DB)
	now := time.Now()

	if input.Status == "approved" {
		// 如果是批准，将临时字段的值更新到正式字段
		_, err := db.Exec(`
			UPDATE provider 
			SET name = COALESCE(temp_name, name),
				icon = COALESCE(temp_icon, icon),
				status = ?,
				updated_at = ?,
				temp_name = NULL,
				temp_icon = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider status"})
			return
		}
	} else {
		// 如果是拒绝，清除临时字段
		_, err := db.Exec(`
			UPDATE provider 
			SET status = ?,
				updated_at = ?,
				temp_name = NULL,
				temp_icon = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if err != nil {
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
	db := c.MustGet("db").(*sql.DB)
	_, err := db.Exec("DELETE FROM provider WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}
