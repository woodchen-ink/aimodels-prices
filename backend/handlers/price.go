package handlers

import (
	"aimodels-prices/database"
	"aimodels-prices/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 在createPrice函数中添加新字段的处理
func createPrice(c *gin.Context) {
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建时间和状态
	price.Status = "pending"
	price.CreatedAt = time.Now()

	// 验证必填字段
	if price.Model == "" || price.ModelType == "" || price.BillingType == "" ||
		price.Currency == "" || price.CreatedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必填字段不能为空"})
		return
	}

	// 验证扩展价格字段（如果提供）
	if price.InputAudioTokens != nil && *price.InputAudioTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "音频输入倍率不能为负数"})
		return
	}
	if price.CachedReadTokens != nil && *price.CachedReadTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缓存读取倍率不能为负数"})
		return
	}
	if price.ReasoningTokens != nil && *price.ReasoningTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "推理倍率不能为负数"})
		return
	}
	if price.InputTextTokens != nil && *price.InputTextTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入文本倍率不能为负数"})
		return
	}
	if price.OutputTextTokens != nil && *price.OutputTextTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输出文本倍率不能为负数"})
		return
	}
	if price.InputImageTokens != nil && *price.InputImageTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入图片倍率不能为负数"})
		return
	}
	if price.OutputImageTokens != nil && *price.OutputImageTokens < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输出图片倍率不能为负数"})
		return
	}

	// 创建价格记录
	if err := database.DB.Create(&price).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, price)
}

// 在updatePrice函数中添加新字段的处理
func updatePrice(c *gin.Context) {
	var price models.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有价格记录
	var existingPrice models.Price
	if err := database.DB.First(&existingPrice, price.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "价格记录不存在"})
		return
	}

	// 更新临时字段
	updates := map[string]interface{}{
		"temp_model":               price.Model,
		"temp_model_type":          price.ModelType,
		"temp_billing_type":        price.BillingType,
		"temp_channel_type":        price.ChannelType,
		"temp_currency":            price.Currency,
		"temp_input_price":         price.InputPrice,
		"temp_output_price":        price.OutputPrice,
		"temp_input_audio_tokens":  price.InputAudioTokens,
		"temp_cached_read_tokens":  price.CachedReadTokens,
		"temp_reasoning_tokens":    price.ReasoningTokens,
		"temp_input_text_tokens":   price.InputTextTokens,
		"temp_output_text_tokens":  price.OutputTextTokens,
		"temp_input_image_tokens":  price.InputImageTokens,
		"temp_output_image_tokens": price.OutputImageTokens,
		"temp_price_source":        price.PriceSource,
		"updated_by":               price.UpdatedBy,
		"status":                   "pending",
	}

	if err := database.DB.Model(&existingPrice).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existingPrice)
}

// 在approvePrice函数中添加新字段的处理
func approvePrice(c *gin.Context) {
	id := c.Param("id")

	var price models.Price
	if err := database.DB.First(&price, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "价格记录不存在"})
		return
	}

	// 更新字段
	updates := map[string]interface{}{
		"model":               price.TempModel,
		"model_type":          price.TempModelType,
		"billing_type":        price.TempBillingType,
		"channel_type":        price.TempChannelType,
		"currency":            price.TempCurrency,
		"input_price":         price.TempInputPrice,
		"output_price":        price.TempOutputPrice,
		"input_audio_tokens":  price.TempInputAudioTokens,
		"cached_read_tokens":  price.TempCachedReadTokens,
		"reasoning_tokens":    price.TempReasoningTokens,
		"input_text_tokens":   price.TempInputTextTokens,
		"output_text_tokens":  price.TempOutputTextTokens,
		"input_image_tokens":  price.TempInputImageTokens,
		"output_image_tokens": price.TempOutputImageTokens,
		"price_source":        price.TempPriceSource,
		"status":              "approved",
		// 清空临时字段
		"temp_model":               nil,
		"temp_model_type":          nil,
		"temp_billing_type":        nil,
		"temp_channel_type":        nil,
		"temp_currency":            nil,
		"temp_input_price":         nil,
		"temp_output_price":        nil,
		"temp_input_audio_tokens":  nil,
		"temp_cached_read_tokens":  nil,
		"temp_reasoning_tokens":    nil,
		"temp_input_text_tokens":   nil,
		"temp_output_text_tokens":  nil,
		"temp_input_image_tokens":  nil,
		"temp_output_image_tokens": nil,
		"temp_price_source":        nil,
		"updated_by":               nil,
	}

	if err := database.DB.Model(&price).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, price)
}
