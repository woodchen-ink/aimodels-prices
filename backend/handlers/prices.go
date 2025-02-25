package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

func GetPrices(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	// 获取分页和筛选参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	channelType := c.Query("channel_type") // 厂商筛选参数
	modelType := c.Query("model_type")     // 模型类型筛选参数

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if channelType != "" {
		conditions = append(conditions, "channel_type = ?")
		args = append(args, channelType)
	}
	if modelType != "" {
		conditions = append(conditions, "model_type = ?")
		args = append(args, modelType)
	}

	// 组合WHERE子句
	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM price"
	if whereClause != "" {
		countQuery += " " + whereClause
	}
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count prices"})
		return
	}

	// 使用分页查询
	query := `
		SELECT id, model, model_type, billing_type, channel_type, currency, input_price, output_price, 
			price_source, status, created_at, updated_at, created_by,
			temp_model, temp_model_type, temp_billing_type, temp_channel_type, temp_currency,
			temp_input_price, temp_output_price, temp_price_source, updated_by
		FROM price`
	if whereClause != "" {
		query += " " + whereClause
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var price models.Price
		if err := rows.Scan(
			&price.ID, &price.Model, &price.ModelType, &price.BillingType, &price.ChannelType, &price.Currency,
			&price.InputPrice, &price.OutputPrice, &price.PriceSource, &price.Status,
			&price.CreatedAt, &price.UpdatedAt, &price.CreatedBy,
			&price.TempModel, &price.TempModelType, &price.TempBillingType, &price.TempChannelType, &price.TempCurrency,
			&price.TempInputPrice, &price.TempOutputPrice, &price.TempPriceSource, &price.UpdatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan price"})
			return
		}
		prices = append(prices, price)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  total,
		"prices": prices,
	})
}

func CreatePrice(c *gin.Context) {
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证模型厂商ID是否存在
	db := c.MustGet("db").(*sql.DB)
	var providerExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM provider WHERE id = ?)", price.ChannelType).Scan(&providerExists)
	if err != nil || !providerExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO price (model, model_type, billing_type, channel_type, currency, input_price, output_price, 
			price_source, status, created_by, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?)`,
		price.Model, price.ModelType, price.BillingType, price.ChannelType, price.Currency,
		price.InputPrice, price.OutputPrice, price.PriceSource, price.CreatedBy,
		now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create price"})
		return
	}

	id, _ := result.LastInsertId()
	price.ID = uint(id)
	price.Status = "pending"
	price.CreatedAt = now
	price.UpdatedAt = now

	c.JSON(http.StatusCreated, price)
}

func UpdatePriceStatus(c *gin.Context) {
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
			UPDATE price 
			SET model = COALESCE(temp_model, model),
				model_type = COALESCE(temp_model_type, model_type),
				billing_type = COALESCE(temp_billing_type, billing_type),
				channel_type = COALESCE(temp_channel_type, channel_type),
				currency = COALESCE(temp_currency, currency),
				input_price = COALESCE(temp_input_price, input_price),
				output_price = COALESCE(temp_output_price, output_price),
				price_source = COALESCE(temp_price_source, price_source),
				status = ?,
				updated_at = ?,
				temp_model = NULL,
				temp_model_type = NULL,
				temp_billing_type = NULL,
				temp_channel_type = NULL,
				temp_currency = NULL,
				temp_input_price = NULL,
				temp_output_price = NULL,
				temp_price_source = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
			return
		}
	} else {
		// 如果是拒绝，清除临时字段
		_, err := db.Exec(`
			UPDATE price 
			SET status = ?,
				updated_at = ?,
				temp_model = NULL,
				temp_model_type = NULL,
				temp_billing_type = NULL,
				temp_channel_type = NULL,
				temp_currency = NULL,
				temp_input_price = NULL,
				temp_output_price = NULL,
				temp_price_source = NULL,
				updated_by = NULL
			WHERE id = ?`, input.Status, now, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Status updated successfully",
		"status":     input.Status,
		"updated_at": now,
	})
}

// UpdatePrice 更新价格
func UpdatePrice(c *gin.Context) {
	id := c.Param("id")
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证模型厂商ID是否存在
	db := c.MustGet("db").(*sql.DB)
	var providerExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM provider WHERE id = ?)", price.ChannelType).Scan(&providerExists)
	if err != nil || !providerExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
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

	var query string
	var args []interface{}

	// 根据用户角色决定更新方式
	if currentUser.Role == "admin" {
		// 管理员直接更新主字段
		query = `
			UPDATE price 
			SET model = ?, model_type = ?, billing_type = ?, channel_type = ?, currency = ?, 
				input_price = ?, output_price = ?, price_source = ?, 
				updated_by = ?, updated_at = ?, status = 'approved',
				temp_model = NULL, temp_model_type = NULL, temp_billing_type = NULL, 
				temp_channel_type = NULL, temp_currency = NULL, temp_input_price = NULL, 
				temp_output_price = NULL, temp_price_source = NULL
			WHERE id = ?`
		args = []interface{}{
			price.Model, price.ModelType, price.BillingType, price.ChannelType, price.Currency,
			price.InputPrice, price.OutputPrice, price.PriceSource,
			currentUser.Username, now, id,
		}
	} else {
		// 普通用户更新临时字段
		query = `
			UPDATE price 
			SET temp_model = ?, temp_model_type = ?, temp_billing_type = ?, temp_channel_type = ?, 
				temp_currency = ?, temp_input_price = ?, temp_output_price = ?, temp_price_source = ?, 
				updated_by = ?, updated_at = ?, status = 'pending'
			WHERE id = ?`
		args = []interface{}{
			price.Model, price.ModelType, price.BillingType, price.ChannelType, price.Currency,
			price.InputPrice, price.OutputPrice, price.PriceSource,
			currentUser.Username, now, id,
		}
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
		return
	}

	// 获取更新后的价格信息
	err = db.QueryRow(`
		SELECT id, model, model_type, billing_type, channel_type, currency, input_price, output_price, 
			price_source, status, created_at, updated_at, created_by,
			temp_model, temp_model_type, temp_billing_type, temp_channel_type, temp_currency,
			temp_input_price, temp_output_price, temp_price_source, updated_by
		FROM price WHERE id = ?`, id).Scan(
		&price.ID, &price.Model, &price.ModelType, &price.BillingType, &price.ChannelType, &price.Currency,
		&price.InputPrice, &price.OutputPrice, &price.PriceSource, &price.Status,
		&price.CreatedAt, &price.UpdatedAt, &price.CreatedBy,
		&price.TempModel, &price.TempModelType, &price.TempBillingType, &price.TempChannelType, &price.TempCurrency,
		&price.TempInputPrice, &price.TempOutputPrice, &price.TempPriceSource, &price.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated price"})
		return
	}

	c.JSON(http.StatusOK, price)
}

// DeletePrice 删除价格
func DeletePrice(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec("DELETE FROM price WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price deleted successfully"})
}

// PriceRate 价格倍率结构
type PriceRate struct {
	Model       string  `json:"model"`
	ModelType   string  `json:"model_type"`
	Type        string  `json:"type"`
	ChannelType uint    `json:"channel_type"`
	Input       float64 `json:"input"`
	Output      float64 `json:"output"`
}

// GetPriceRates 获取价格倍率
func GetPriceRates(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query(`
		SELECT model, model_type, billing_type, channel_type, 
			CASE 
				WHEN currency = 'USD' THEN input_price / 2
				ELSE input_price / 14
			END as input_rate,
			CASE 
				WHEN currency = 'USD' THEN output_price / 2
				ELSE output_price / 14
			END as output_rate
		FROM price 
		WHERE status = 'approved'
		ORDER BY model, channel_type`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price rates"})
		return
	}
	defer rows.Close()

	var rates []PriceRate
	for rows.Next() {
		var rate PriceRate
		if err := rows.Scan(
			&rate.Model,
			&rate.ModelType,
			&rate.Type,
			&rate.ChannelType,
			&rate.Input,
			&rate.Output); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan price rate"})
			return
		}
		rates = append(rates, rate)
	}

	c.JSON(http.StatusOK, rates)
}
