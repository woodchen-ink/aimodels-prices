package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/handlers/one_hub"
	"aimodels-prices/models"
)

func GetPrices(c *gin.Context) {
	// 获取分页和筛选参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	channelType := c.Query("channel_type") // 厂商筛选参数
	modelType := c.Query("model_type")     // 模型类型筛选参数
	searchQuery := c.Query("search")       // 搜索查询参数
	status := c.Query("status")            // 状态筛选参数

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 构建缓存键
	cacheKey := fmt.Sprintf("prices_page_%d_size_%d_channel_%s_type_%s_search_%s_status_%s",
		page, pageSize, channelType, modelType, searchQuery, status)

	// 尝试从缓存获取
	if cachedData, found := database.GlobalCache.Get(cacheKey); found {
		if result, ok := cachedData.(gin.H); ok {
			c.JSON(http.StatusOK, result)
			return
		}
	}

	// 构建查询 - 使用索引优化
	query := database.DB.Model(&models.Price{}).Select("*")

	// 添加筛选条件
	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	if modelType != "" {
		query = query.Where("model_type = ?", modelType)
	}
	// 添加搜索条件
	if searchQuery != "" {
		query = query.Where("model LIKE ?", "%"+searchQuery+"%")
	}
	// 添加状态筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数 - 使用缓存优化
	var total int64
	totalCacheKey := fmt.Sprintf("prices_count_channel_%s_type_%s_search_%s_status_%s",
		channelType, modelType, searchQuery, status)

	if cachedTotal, found := database.GlobalCache.Get(totalCacheKey); found {
		if t, ok := cachedTotal.(int64); ok {
			total = t
		} else {
			if err := query.Count(&total).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count prices"})
				return
			}
			database.GlobalCache.Set(totalCacheKey, total, 5*time.Minute)
		}
	} else {
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count prices"})
			return
		}
		database.GlobalCache.Set(totalCacheKey, total, 5*time.Minute)
	}

	// 获取分页数据 - 使用索引优化
	var prices []models.Price
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}

	result := gin.H{
		"total": total,
		"data":  prices,
	}

	// 存入缓存，有效期5分钟
	database.GlobalCache.Set(cacheKey, result, 5*time.Minute)

	c.JSON(http.StatusOK, result)
}

// processPrice 处理价格的创建和更新逻辑,只负责处理业务逻辑
func ProcessPrice(price models.Price, existingPrice *models.Price, isAdmin bool, username string) (models.Price, bool, error) {
	// 如果是更新操作且存在现有记录
	if existingPrice != nil {
		// 使用更精确的浮点数比较函数
		priceEqual := func(a, b float64) bool {
			// 使用epsilon值进行浮点数比较，考虑到价格通常精确到小数点后4位
			epsilon := 0.00001
			return math.Abs(a-b) < epsilon
		}

		// 比较指针类型的浮点数是否相等
		pointerPriceEqual := func(a, b *float64) bool {
			if a == nil && b == nil {
				return true
			}
			if a == nil || b == nil {
				return false
			}
			return priceEqual(*a, *b)
		}

		// 检查价格是否有变化
		if isAdmin {
			// 管理员直接更新主字段，检查是否有实际变化
			if existingPrice.Model == price.Model &&
				existingPrice.ModelType == price.ModelType &&
				existingPrice.BillingType == price.BillingType &&
				existingPrice.ChannelType == price.ChannelType &&
				existingPrice.Currency == price.Currency &&
				priceEqual(existingPrice.InputPrice, price.InputPrice) &&
				priceEqual(existingPrice.OutputPrice, price.OutputPrice) &&
				pointerPriceEqual(existingPrice.InputAudioTokens, price.InputAudioTokens) &&
				pointerPriceEqual(existingPrice.OutputAudioTokens, price.OutputAudioTokens) &&
				pointerPriceEqual(existingPrice.CachedTokens, price.CachedTokens) &&
				pointerPriceEqual(existingPrice.CachedReadTokens, price.CachedReadTokens) &&
				pointerPriceEqual(existingPrice.CachedWriteTokens, price.CachedWriteTokens) &&
				pointerPriceEqual(existingPrice.ReasoningTokens, price.ReasoningTokens) &&
				pointerPriceEqual(existingPrice.InputTextTokens, price.InputTextTokens) &&
				pointerPriceEqual(existingPrice.OutputTextTokens, price.OutputTextTokens) &&
				pointerPriceEqual(existingPrice.InputImageTokens, price.InputImageTokens) &&
				pointerPriceEqual(existingPrice.OutputImageTokens, price.OutputImageTokens) &&
				existingPrice.PriceSource == price.PriceSource {
				// 没有变化，不需要更新
				return *existingPrice, false, nil
			}

			// 有变化，更新字段
			existingPrice.Model = price.Model
			existingPrice.ModelType = price.ModelType
			existingPrice.BillingType = price.BillingType
			existingPrice.ChannelType = price.ChannelType
			existingPrice.Currency = price.Currency
			existingPrice.InputPrice = price.InputPrice
			existingPrice.OutputPrice = price.OutputPrice
			existingPrice.InputAudioTokens = price.InputAudioTokens
			existingPrice.OutputAudioTokens = price.OutputAudioTokens
			existingPrice.CachedTokens = price.CachedTokens
			existingPrice.CachedReadTokens = price.CachedReadTokens
			existingPrice.CachedWriteTokens = price.CachedWriteTokens
			existingPrice.ReasoningTokens = price.ReasoningTokens
			existingPrice.InputTextTokens = price.InputTextTokens
			existingPrice.OutputTextTokens = price.OutputTextTokens
			existingPrice.InputImageTokens = price.InputImageTokens
			existingPrice.OutputImageTokens = price.OutputImageTokens
			existingPrice.PriceSource = price.PriceSource
			existingPrice.Status = "approved"
			existingPrice.UpdatedBy = &username
			existingPrice.TempModel = nil
			existingPrice.TempModelType = nil
			existingPrice.TempBillingType = nil
			existingPrice.TempChannelType = nil
			existingPrice.TempCurrency = nil
			existingPrice.TempInputPrice = nil
			existingPrice.TempOutputPrice = nil
			existingPrice.TempInputAudioTokens = nil
			existingPrice.TempOutputAudioTokens = nil
			existingPrice.TempCachedTokens = nil
			existingPrice.TempCachedReadTokens = nil
			existingPrice.TempCachedWriteTokens = nil
			existingPrice.TempReasoningTokens = nil
			existingPrice.TempInputTextTokens = nil
			existingPrice.TempOutputTextTokens = nil
			existingPrice.TempInputImageTokens = nil
			existingPrice.TempOutputImageTokens = nil
			existingPrice.TempPriceSource = nil

			// 保存更新
			if err := database.DB.Save(existingPrice).Error; err != nil {
				return *existingPrice, false, err
			}
			return *existingPrice, true, nil
		} else {
			// 普通用户更新临时字段，检查是否有实际变化

			// 先检查与主字段比较是否有变化
			hasChanges := false

			if existingPrice.Model != price.Model ||
				existingPrice.ModelType != price.ModelType ||
				existingPrice.BillingType != price.BillingType ||
				existingPrice.ChannelType != price.ChannelType ||
				existingPrice.Currency != price.Currency ||
				!priceEqual(existingPrice.InputPrice, price.InputPrice) ||
				!priceEqual(existingPrice.OutputPrice, price.OutputPrice) ||
				!pointerPriceEqual(existingPrice.InputAudioTokens, price.InputAudioTokens) ||
				!pointerPriceEqual(existingPrice.OutputAudioTokens, price.OutputAudioTokens) ||
				!pointerPriceEqual(existingPrice.CachedTokens, price.CachedTokens) ||
				!pointerPriceEqual(existingPrice.CachedReadTokens, price.CachedReadTokens) ||
				!pointerPriceEqual(existingPrice.CachedWriteTokens, price.CachedWriteTokens) ||
				!pointerPriceEqual(existingPrice.ReasoningTokens, price.ReasoningTokens) ||
				!pointerPriceEqual(existingPrice.InputTextTokens, price.InputTextTokens) ||
				!pointerPriceEqual(existingPrice.OutputTextTokens, price.OutputTextTokens) ||
				!pointerPriceEqual(existingPrice.InputImageTokens, price.InputImageTokens) ||
				!pointerPriceEqual(existingPrice.OutputImageTokens, price.OutputImageTokens) ||
				existingPrice.PriceSource != price.PriceSource {
				hasChanges = true
			}

			// 如果与主字段有变化，再检查与临时字段比较是否有变化
			if hasChanges && existingPrice.TempModel != nil {
				// 检查是否与已有的临时字段相同
				if *existingPrice.TempModel == price.Model &&
					(existingPrice.TempModelType == nil || *existingPrice.TempModelType == price.ModelType) &&
					(existingPrice.TempBillingType == nil || *existingPrice.TempBillingType == price.BillingType) &&
					(existingPrice.TempChannelType == nil || *existingPrice.TempChannelType == price.ChannelType) &&
					(existingPrice.TempCurrency == nil || *existingPrice.TempCurrency == price.Currency) &&
					(existingPrice.TempInputPrice == nil || priceEqual(*existingPrice.TempInputPrice, price.InputPrice)) &&
					(existingPrice.TempOutputPrice == nil || priceEqual(*existingPrice.TempOutputPrice, price.OutputPrice)) &&
					(existingPrice.TempInputAudioTokens == nil || pointerPriceEqual(existingPrice.TempInputAudioTokens, price.InputAudioTokens)) &&
					(existingPrice.TempOutputAudioTokens == nil || pointerPriceEqual(existingPrice.TempOutputAudioTokens, price.OutputAudioTokens)) &&
					(existingPrice.TempCachedTokens == nil || pointerPriceEqual(existingPrice.TempCachedTokens, price.CachedTokens)) &&
					(existingPrice.TempCachedReadTokens == nil || pointerPriceEqual(existingPrice.TempCachedReadTokens, price.CachedReadTokens)) &&
					(existingPrice.TempCachedWriteTokens == nil || pointerPriceEqual(existingPrice.TempCachedWriteTokens, price.CachedWriteTokens)) &&
					(existingPrice.TempReasoningTokens == nil || pointerPriceEqual(existingPrice.TempReasoningTokens, price.ReasoningTokens)) &&
					(existingPrice.TempInputTextTokens == nil || pointerPriceEqual(existingPrice.TempInputTextTokens, price.InputTextTokens)) &&
					(existingPrice.TempOutputTextTokens == nil || pointerPriceEqual(existingPrice.TempOutputTextTokens, price.OutputTextTokens)) &&
					(existingPrice.TempInputImageTokens == nil || pointerPriceEqual(existingPrice.TempInputImageTokens, price.InputImageTokens)) &&
					(existingPrice.TempOutputImageTokens == nil || pointerPriceEqual(existingPrice.TempOutputImageTokens, price.OutputImageTokens)) &&
					(existingPrice.TempPriceSource == nil || *existingPrice.TempPriceSource == price.PriceSource) {
					// 与之前提交的临时值相同，不需要更新
					hasChanges = false
				}
			}

			// 如果没有实际变化，直接返回
			if !hasChanges {
				return *existingPrice, false, nil
			}

			// 有变化，更新临时字段
			existingPrice.TempModel = &price.Model
			existingPrice.TempModelType = &price.ModelType
			existingPrice.TempBillingType = &price.BillingType
			existingPrice.TempChannelType = &price.ChannelType
			existingPrice.TempCurrency = &price.Currency
			existingPrice.TempInputPrice = &price.InputPrice
			existingPrice.TempOutputPrice = &price.OutputPrice
			existingPrice.TempInputAudioTokens = price.InputAudioTokens
			existingPrice.TempOutputAudioTokens = price.OutputAudioTokens
			existingPrice.TempCachedTokens = price.CachedTokens
			existingPrice.TempCachedReadTokens = price.CachedReadTokens
			existingPrice.TempCachedWriteTokens = price.CachedWriteTokens
			existingPrice.TempReasoningTokens = price.ReasoningTokens
			existingPrice.TempInputTextTokens = price.InputTextTokens
			existingPrice.TempOutputTextTokens = price.OutputTextTokens
			existingPrice.TempInputImageTokens = price.InputImageTokens
			existingPrice.TempOutputImageTokens = price.OutputImageTokens
			existingPrice.TempPriceSource = &price.PriceSource
			existingPrice.Status = "pending"
			existingPrice.UpdatedBy = &username

			// 保存更新
			if err := database.DB.Save(existingPrice).Error; err != nil {
				return *existingPrice, false, err
			}
			return *existingPrice, true, nil
		}
	} else {
		// 创建新记录
		price.Status = "pending"
		if isAdmin {
			price.Status = "approved"
		}
		price.CreatedBy = username

		// 验证扩展价格字段（如果提供）
		if price.InputAudioTokens != nil && *price.InputAudioTokens < 0 {
			return price, false, fmt.Errorf("音频输入价格不能为负数")
		}
		if price.OutputAudioTokens != nil && *price.OutputAudioTokens < 0 {
			return price, false, fmt.Errorf("音频输出价格不能为负数")
		}
		if price.CachedTokens != nil && *price.CachedTokens < 0 {
			return price, false, fmt.Errorf("缓存价格不能为负数")
		}
		if price.CachedReadTokens != nil && *price.CachedReadTokens < 0 {
			return price, false, fmt.Errorf("缓存读取价格不能为负数")
		}
		if price.CachedWriteTokens != nil && *price.CachedWriteTokens < 0 {
			return price, false, fmt.Errorf("缓存写入价格不能为负数")
		}
		if price.ReasoningTokens != nil && *price.ReasoningTokens < 0 {
			return price, false, fmt.Errorf("推理价格不能为负数")
		}
		if price.InputTextTokens != nil && *price.InputTextTokens < 0 {
			return price, false, fmt.Errorf("输入文本价格不能为负数")
		}
		if price.OutputTextTokens != nil && *price.OutputTextTokens < 0 {
			return price, false, fmt.Errorf("输出文本价格不能为负数")
		}
		if price.InputImageTokens != nil && *price.InputImageTokens < 0 {
			return price, false, fmt.Errorf("输入图片价格不能为负数")
		}
		if price.OutputImageTokens != nil && *price.OutputImageTokens < 0 {
			return price, false, fmt.Errorf("输出图片价格不能为负数")
		}

		// 保存新记录
		if err := database.DB.Create(&price).Error; err != nil {
			return price, false, err
		}
		return price, true, nil
	}
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

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	currentUser := user.(*models.User)

	// 处理价格创建
	result, changed, err := ProcessPrice(price, nil, currentUser.Role == "admin", currentUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create price"})
		return
	}

	// 清除所有价格相关缓存
	if changed {
		clearPriceCache()
	}

	c.JSON(http.StatusCreated, result)
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
		if price.TempInputAudioTokens != nil {
			updateMap["input_audio_tokens"] = *price.TempInputAudioTokens
		}
		if price.TempOutputAudioTokens != nil {
			updateMap["output_audio_tokens"] = *price.TempOutputAudioTokens
		}
		if price.TempCachedTokens != nil {
			updateMap["cached_tokens"] = *price.TempCachedTokens
		}
		if price.TempCachedReadTokens != nil {
			updateMap["cached_read_tokens"] = *price.TempCachedReadTokens
		}
		if price.TempCachedWriteTokens != nil {
			updateMap["cached_write_tokens"] = *price.TempCachedWriteTokens
		}
		if price.TempReasoningTokens != nil {
			updateMap["reasoning_tokens"] = *price.TempReasoningTokens
		}
		if price.TempInputTextTokens != nil {
			updateMap["input_text_tokens"] = *price.TempInputTextTokens
		}
		if price.TempOutputTextTokens != nil {
			updateMap["output_text_tokens"] = *price.TempOutputTextTokens
		}
		if price.TempInputImageTokens != nil {
			updateMap["input_image_tokens"] = *price.TempInputImageTokens
		}
		if price.TempOutputImageTokens != nil {
			updateMap["output_image_tokens"] = *price.TempOutputImageTokens
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
		updateMap["temp_input_audio_tokens"] = nil
		updateMap["temp_output_audio_tokens"] = nil
		updateMap["temp_cached_tokens"] = nil
		updateMap["temp_cached_read_tokens"] = nil
		updateMap["temp_cached_write_tokens"] = nil
		updateMap["temp_reasoning_tokens"] = nil
		updateMap["temp_input_text_tokens"] = nil
		updateMap["temp_output_text_tokens"] = nil
		updateMap["temp_input_image_tokens"] = nil
		updateMap["temp_output_image_tokens"] = nil
		updateMap["temp_price_source"] = nil

		if err := tx.Model(&price).Updates(updateMap).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
			return
		}
	} else {
		// 如果是拒绝
		// 检查是否是新创建的价格（没有原始价格）
		isNewPrice := price.Model == "" || (price.TempModel != nil && price.Model == *price.TempModel)

		if isNewPrice {
			// 如果是新创建的价格，直接删除
			if err := tx.Delete(&price).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rejected price"})
				return
			}
		} else {
			// 如果是更新的价格，恢复到原始状态（清除临时字段并设置状态为approved）
			if err := tx.Model(&price).Updates(map[string]interface{}{
				"status":                   "approved", // 恢复为已批准状态
				"updated_at":               time.Now(),
				"temp_model":               nil,
				"temp_model_type":          nil,
				"temp_billing_type":        nil,
				"temp_channel_type":        nil,
				"temp_currency":            nil,
				"temp_input_price":         nil,
				"temp_output_price":        nil,
				"temp_input_audio_tokens":  nil,
				"temp_output_audio_tokens": nil,
				"temp_cached_tokens":       nil,
				"temp_cached_read_tokens":  nil,
				"temp_cached_write_tokens": nil,
				"temp_reasoning_tokens":    nil,
				"temp_input_text_tokens":   nil,
				"temp_output_text_tokens":  nil,
				"temp_input_image_tokens":  nil,
				"temp_output_image_tokens": nil,
				"temp_price_source":        nil,
				"updated_by":               nil,
			}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price status"})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// 清除所有价格相关缓存
	clearPriceCache()

	// 根据操作类型返回不同的消息
	if input.Status == "rejected" && (price.Model == "" || (price.TempModel != nil && price.Model == *price.TempModel)) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Price rejected and deleted successfully",
			"status":     input.Status,
			"updated_at": time.Now(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Status updated successfully",
			"status":     input.Status,
			"updated_at": time.Now(),
		})
	}
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

	// 处理价格更新
	result, changed, err := ProcessPrice(price, &existingPrice, currentUser.Role == "admin", currentUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
		return
	}

	// 清除所有价格相关缓存
	if changed {
		clearPriceCache()
	}

	c.JSON(http.StatusOK, result)
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

	// 清除所有价格相关缓存
	clearPriceCache()

	c.JSON(http.StatusOK, gin.H{"message": "Price deleted successfully"})
}

func ApproveAllPrices(c *gin.Context) {
	// 获取操作类型（批准或拒绝）
	var input struct {
		Action string `json:"action" binding:"required,oneof=approve reject"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action, must be 'approve' or 'reject'"})
		return
	}

	// 查找所有待审核的价格
	var pendingPrices []models.Price
	if err := database.DB.Where("status = 'pending'").Find(&pendingPrices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending prices"})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	processedCount := 0
	deletedCount := 0

	for _, price := range pendingPrices {
		if input.Action == "approve" {
			// 批准操作
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
			if price.TempInputAudioTokens != nil {
				updateMap["input_audio_tokens"] = *price.TempInputAudioTokens
			}
			if price.TempOutputAudioTokens != nil {
				updateMap["output_audio_tokens"] = *price.TempOutputAudioTokens
			}
			if price.TempCachedTokens != nil {
				updateMap["cached_tokens"] = *price.TempCachedTokens
			}
			if price.TempCachedReadTokens != nil {
				updateMap["cached_read_tokens"] = *price.TempCachedReadTokens
			}
			if price.TempCachedWriteTokens != nil {
				updateMap["cached_write_tokens"] = *price.TempCachedWriteTokens
			}
			if price.TempReasoningTokens != nil {
				updateMap["reasoning_tokens"] = *price.TempReasoningTokens
			}
			if price.TempInputTextTokens != nil {
				updateMap["input_text_tokens"] = *price.TempInputTextTokens
			}
			if price.TempOutputTextTokens != nil {
				updateMap["output_text_tokens"] = *price.TempOutputTextTokens
			}
			if price.TempInputImageTokens != nil {
				updateMap["input_image_tokens"] = *price.TempInputImageTokens
			}
			if price.TempOutputImageTokens != nil {
				updateMap["output_image_tokens"] = *price.TempOutputImageTokens
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
			updateMap["temp_input_audio_tokens"] = nil
			updateMap["temp_output_audio_tokens"] = nil
			updateMap["temp_cached_tokens"] = nil
			updateMap["temp_cached_read_tokens"] = nil
			updateMap["temp_cached_write_tokens"] = nil
			updateMap["temp_reasoning_tokens"] = nil
			updateMap["temp_input_text_tokens"] = nil
			updateMap["temp_output_text_tokens"] = nil
			updateMap["temp_input_image_tokens"] = nil
			updateMap["temp_output_image_tokens"] = nil
			updateMap["temp_price_source"] = nil

			if err := tx.Model(&price).Updates(updateMap).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve prices"})
				return
			}
			processedCount++
		} else {
			// 拒绝操作
			// 检查是否是新创建的价格（没有原始价格）
			isNewPrice := price.Model == "" || (price.TempModel != nil && price.Model == *price.TempModel)

			if isNewPrice {
				// 如果是新创建的价格，直接删除
				if err := tx.Delete(&price).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rejected price"})
					return
				}
				deletedCount++
			} else {
				// 如果是更新的价格，恢复到原始状态（清除临时字段并设置状态为approved）
				if err := tx.Model(&price).Updates(map[string]interface{}{
					"status":                   "approved", // 恢复为已批准状态
					"updated_at":               time.Now(),
					"temp_model":               nil,
					"temp_model_type":          nil,
					"temp_billing_type":        nil,
					"temp_channel_type":        nil,
					"temp_currency":            nil,
					"temp_input_price":         nil,
					"temp_output_price":        nil,
					"temp_input_audio_tokens":  nil,
					"temp_output_audio_tokens": nil,
					"temp_cached_tokens":       nil,
					"temp_cached_read_tokens":  nil,
					"temp_cached_write_tokens": nil,
					"temp_reasoning_tokens":    nil,
					"temp_input_text_tokens":   nil,
					"temp_output_text_tokens":  nil,
					"temp_input_image_tokens":  nil,
					"temp_output_image_tokens": nil,
					"temp_price_source":        nil,
					"updated_by":               nil,
				}).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject prices"})
					return
				}
				processedCount++
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// 清除所有价格相关缓存
	clearPriceCache()

	// 根据操作类型返回不同的消息
	if input.Action == "approve" {
		c.JSON(http.StatusOK, gin.H{
			"message": "All pending prices approved successfully",
			"count":   processedCount,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "All pending prices rejected successfully",
			"processed": processedCount,
			"deleted":   deletedCount,
			"total":     processedCount + deletedCount,
		})
	}
}

// clearPriceCache 清除所有价格相关的缓存
func clearPriceCache() {
	// 由于我们无法精确知道哪些缓存键与价格相关，所以清除所有缓存
	database.GlobalCache.Clear()

	// 同时清除价格倍率缓存
	one_hub.ClearRatesCache()
}
