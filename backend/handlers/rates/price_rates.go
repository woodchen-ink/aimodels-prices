package rates

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// PriceRate 价格倍率结构
type PriceRate struct {
	Model       string  `json:"model"`
	Type        string  `json:"type"`
	ChannelType uint    `json:"channel_type"`
	Input       float64 `json:"input"`
	Output      float64 `json:"output"`
}

// GetPriceRates 获取价格倍率
func GetPriceRates(c *gin.Context) {
	cacheKey := "price_rates"

	// 尝试从缓存获取
	if cachedData, found := database.GlobalCache.Get(cacheKey); found {
		if rates, ok := cachedData.([]PriceRate); ok {
			c.JSON(http.StatusOK, rates)
			return
		}
	}

	// 使用索引优化查询，只查询需要的字段
	var prices []models.Price
	if err := database.DB.Select("model, billing_type, channel_type, input_price, output_price, currency, status").
		Where("status = 'approved'").
		Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}

	// 预分配rates切片，减少内存分配
	rates := make([]PriceRate, 0, len(prices))

	// 计算倍率
	for _, price := range prices {
		// 根据货币类型计算倍率
		var inputRate, outputRate float64

		if price.Currency == "USD" {
			// 如果是美元，除以2
			inputRate = price.InputPrice / 2
			outputRate = price.OutputPrice / 2
		} else {
			// 如果是人民币或其他货币，除以14
			inputRate = price.InputPrice / 14
			outputRate = price.OutputPrice / 14
		}

		rates = append(rates, PriceRate{
			Model:       price.Model,
			Type:        price.BillingType,
			ChannelType: price.ChannelType,
			Input:       inputRate,
			Output:      outputRate,
		})
	}

	// 存入缓存，有效期24小时
	database.GlobalCache.Set(cacheKey, rates, 24*time.Hour)

	c.JSON(http.StatusOK, rates)
}

// ClearRatesCache 清除价格倍率缓存
func ClearRatesCache() {
	database.GlobalCache.Delete("price_rates")
}
