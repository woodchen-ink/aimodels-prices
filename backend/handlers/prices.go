package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

func GetPrices(c *gin.Context) {
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

	// 构建查询
	query := database.DB.Model(&models.Price{})

	// 添加筛选条件
	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	if modelType != "" {
		query = query.Where("model_type = ?", modelType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count prices"})
		return
	}

	// 获取分页数据
	var prices []models.Price
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  prices,
	})
}

func CreatePrice(c *gin.Context) {
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证模型厂商ID是否存在
	var provider models.Provider
	if err := database.DB.Where("id = ?", price.ChannelType).First(&provider).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	// 检查同一厂商下是否已存在相同名称的模型
	var count int64
	if err := database.DB.Model(&models.Price{}).Where("channel_type = ? AND model = ? AND status = 'approved'",
		price.ChannelType, price.Model).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check model existence"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model with the same name already exists for this provider"})
		return
	}

	// 设置状态和创建者
	price.Status = "pending"

	// 创建记录
	if err := database.DB.Create(&price).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create price"})
		return
	}

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

	// 查找价格记录
	var price models.Price
	if err := database.DB.Where("id = ?", id).First(&price).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Price not found"})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	if input.Status == "approved" {
		// 如果是批准，将临时字段的值更新到正式字段
		updateMap := map[string]interface{}{
			"status":     input.Status,
			"updated_at": time.Now(),
		}

		// 如果临时字段有值，则更新主字段
		if price.TempModel != nil {
			updateMap["model"] = *price.TempModel
		}
		if price.TempModelType != nil {
			updateMap["model_type"] = *price.TempModelType
		}
		if price.TempBillingType != nil {
			updateMap["billing_type"] = *price.TempBillingType
		}
		if price.TempChannelType != nil {
			updateMap["channel_type"] = *price.TempChannelType
		}
		if price.TempCurrency != nil {
			updateMap["currency"] = *price.TempCurrency
		}
		if price.TempInputPrice != nil {
			updateMap["input_price"] = *price.TempInputPrice
		}
		if price.TempOutputPrice != nil {
			updateMap["output_price"] = *price.TempOutputPrice
		}
		if price.TempPriceSource != nil {
			updateMap["price_source"] = *price.TempPriceSource
		}

		// 清除所有临时字段
		updateMap["temp_model"] = nil
		updateMap["temp_model_type"] = nil
		updateMap["temp_billing_type"] = nil
		updateMap["temp_channel_type"] = nil
		updateMap["temp_currency"] = nil
		updateMap["temp_input_price"] = nil
		updateMap["temp_output_price"] = nil
		updateMap["temp_price_source"] = nil
		updateMap["updated_by"] = nil

		if err := tx.Model(&price).Updates(updateMap).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
			return
		}
	} else {
		// 如果是拒绝，清除临时字段
		if err := tx.Model(&price).Updates(map[string]interface{}{
			"status":            input.Status,
			"updated_at":        time.Now(),
			"temp_model":        nil,
			"temp_model_type":   nil,
			"temp_billing_type": nil,
			"temp_channel_type": nil,
			"temp_currency":     nil,
			"temp_input_price":  nil,
			"temp_output_price": nil,
			"temp_price_source": nil,
			"updated_by":        nil,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Status updated successfully",
		"status":     input.Status,
		"updated_at": time.Now(),
	})
}

func UpdatePrice(c *gin.Context) {
	id := c.Param("id")
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证模型厂商ID是否存在
	var provider models.Provider
	if err := database.DB.Where("id = ?", price.ChannelType).First(&provider).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	// 检查同一厂商下是否已存在相同名称的模型（排除当前正在编辑的记录）
	var count int64
	if err := database.DB.Model(&models.Price{}).Where("channel_type = ? AND model = ? AND id != ? AND status = 'approved'",
		price.ChannelType, price.Model, id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check model existence"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model with the same name already exists for this provider"})
		return
	}

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	currentUser := user.(*models.User)

	// 查找现有记录
	var existingPrice models.Price
	if err := database.DB.Where("id = ?", id).First(&existingPrice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Price not found"})
		return
	}

	// 根据用户角色决定更新方式
	if currentUser.Role == "admin" {
		// 管理员直接更新主字段
		existingPrice.Model = price.Model
		existingPrice.ModelType = price.ModelType
		existingPrice.BillingType = price.BillingType
		existingPrice.ChannelType = price.ChannelType
		existingPrice.Currency = price.Currency
		existingPrice.InputPrice = price.InputPrice
		existingPrice.OutputPrice = price.OutputPrice
		existingPrice.PriceSource = price.PriceSource
		existingPrice.Status = "approved"
		existingPrice.UpdatedBy = &currentUser.Username
		existingPrice.TempModel = nil
		existingPrice.TempModelType = nil
		existingPrice.TempBillingType = nil
		existingPrice.TempChannelType = nil
		existingPrice.TempCurrency = nil
		existingPrice.TempInputPrice = nil
		existingPrice.TempOutputPrice = nil
		existingPrice.TempPriceSource = nil

		if err := database.DB.Save(&existingPrice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
			return
		}
	} else {
		// 普通用户更新临时字段
		existingPrice.TempModel = &price.Model
		existingPrice.TempModelType = &price.ModelType
		existingPrice.TempBillingType = &price.BillingType
		existingPrice.TempChannelType = &price.ChannelType
		existingPrice.TempCurrency = &price.Currency
		existingPrice.TempInputPrice = &price.InputPrice
		existingPrice.TempOutputPrice = &price.OutputPrice
		existingPrice.TempPriceSource = &price.PriceSource
		existingPrice.Status = "pending"
		existingPrice.UpdatedBy = &currentUser.Username

		if err := database.DB.Save(&existingPrice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
			return
		}
	}

	c.JSON(http.StatusOK, existingPrice)
}

func DeletePrice(c *gin.Context) {
	id := c.Param("id")

	// 查找价格记录
	var price models.Price
	if err := database.DB.Where("id = ?", id).First(&price).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Price not found"})
		return
	}

	// 删除记录
	if err := database.DB.Delete(&price).Error; err != nil {
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
	var prices []models.Price
	if err := database.DB.Where("status = 'approved'").Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}

	// 按模型分组
	modelMap := make(map[string]map[uint]models.Price)
	for _, price := range prices {
		if _, exists := modelMap[price.Model]; !exists {
			modelMap[price.Model] = make(map[uint]models.Price)
		}
		modelMap[price.Model][price.ChannelType] = price
	}

	// 计算倍率
	var rates []PriceRate
	for model, providers := range modelMap {
		// 找出基准价格（通常是OpenAI的价格）
		var basePrice models.Price
		var found bool
		for _, price := range providers {
			if price.ChannelType == 1 { // 假设OpenAI的ID是1
				basePrice = price
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// 计算其他厂商相对于基准价格的倍率
		for channelType, price := range providers {
			if channelType == 1 {
				continue // 跳过基准价格
			}

			// 计算输入和输出的倍率
			inputRate := 0.0
			if basePrice.InputPrice > 0 {
				inputRate = price.InputPrice / basePrice.InputPrice
			}

			outputRate := 0.0
			if basePrice.OutputPrice > 0 {
				outputRate = price.OutputPrice / basePrice.OutputPrice
			}

			rates = append(rates, PriceRate{
				Model:       model,
				ModelType:   price.ModelType,
				Type:        price.BillingType,
				ChannelType: channelType,
				Input:       inputRate,
				Output:      outputRate,
			})
		}
	}

	c.JSON(http.StatusOK, rates)
}

func ApproveAllPrices(c *gin.Context) {
	// 查找所有待审核的价格
	var pendingPrices []models.Price
	if err := database.DB.Where("status = 'pending'").Find(&pendingPrices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending prices"})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	for _, price := range pendingPrices {
		updateMap := map[string]interface{}{
			"status":     "approved",
			"updated_at": time.Now(),
		}

		// 如果临时字段有值，则更新主字段
		if price.TempModel != nil {
			updateMap["model"] = *price.TempModel
		}
		if price.TempModelType != nil {
			updateMap["model_type"] = *price.TempModelType
		}
		if price.TempBillingType != nil {
			updateMap["billing_type"] = *price.TempBillingType
		}
		if price.TempChannelType != nil {
			updateMap["channel_type"] = *price.TempChannelType
		}
		if price.TempCurrency != nil {
			updateMap["currency"] = *price.TempCurrency
		}
		if price.TempInputPrice != nil {
			updateMap["input_price"] = *price.TempInputPrice
		}
		if price.TempOutputPrice != nil {
			updateMap["output_price"] = *price.TempOutputPrice
		}
		if price.TempPriceSource != nil {
			updateMap["price_source"] = *price.TempPriceSource
		}

		// 清除所有临时字段
		updateMap["temp_model"] = nil
		updateMap["temp_model_type"] = nil
		updateMap["temp_billing_type"] = nil
		updateMap["temp_channel_type"] = nil
		updateMap["temp_currency"] = nil
		updateMap["temp_input_price"] = nil
		updateMap["temp_output_price"] = nil
		updateMap["temp_price_source"] = nil
		updateMap["updated_by"] = nil

		if err := tx.Model(&price).Updates(updateMap).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve prices"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All pending prices approved successfully",
		"count":   len(pendingPrices),
	})
}
