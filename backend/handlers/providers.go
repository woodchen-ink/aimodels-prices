package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

// GetProviders 获取所有供应商
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

// CreateProvider 创建供应商
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

// UpdateProvider 更新供应商
func UpdateProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var provider models.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	now := time.Now()
	_, err = db.Exec(`
		UPDATE provider 
		SET name = ?, icon = ?, updated_at = ?
		WHERE id = ?`,
		provider.Name, provider.Icon, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	// 获取更新后的供应商信息
	err = db.QueryRow(`
		SELECT id, name, icon, created_at, updated_at, created_by
		FROM provider WHERE id = ?`, id).Scan(
		&provider.ID, &provider.Name, &provider.Icon,
		&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// UpdateProviderStatus 更新供应商状态
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

// DeleteProvider 删除供应商
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
