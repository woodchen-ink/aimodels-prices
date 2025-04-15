package rates

import (
	"net/http"
	"strings"
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

	// 创建map用于存储模型及其对应的最高倍率
	modelRateMap := make(map[string]PriceRate)

	// 计算倍率
	for _, price := range prices {
		// 根据货币类型计算倍率
		var inputRate, outputRate float64

		if price.Currency == "USD" {
			// 如果是美元，除以2
			inputRate = round(price.InputPrice/2, 4)
			outputRate = round(price.OutputPrice/2, 4)
		} else {
			// 如果是人民币或其他货币，除以14
			inputRate = round(price.InputPrice/14, 4)
			outputRate = round(price.OutputPrice/14, 4)
		}

		// 创建当前价格的PriceRate
		currentRate := PriceRate{
			Model:       price.Model,
			Type:        price.BillingType,
			ChannelType: price.ChannelType,
			Input:       inputRate,
			Output:      outputRate,
		}

		// 转换为小写以实现不区分大小写比较
		modelLower := strings.ToLower(price.Model)

		// 检查是否已存在相同模型名称（不区分大小写）
		if existingRate, exists := modelRateMap[modelLower]; exists {
			// 比较倍率，保留较高的那个
			// 这里我们以输入和输出倍率的总和作为比较标准
			existingTotal := existingRate.Input + existingRate.Output
			currentTotal := inputRate + outputRate

			if currentTotal > existingTotal {
				// 当前倍率更高，替换已存在的
				modelRateMap[modelLower] = currentRate
			}
		} else {
			// 不存在相同模型名称，直接添加
			modelRateMap[modelLower] = currentRate
		}
	}

	// 从map中提取结果到slice
	rates := make([]PriceRate, 0, len(modelRateMap))
	for _, rate := range modelRateMap {
		rates = append(rates, rate)
	}

	// 存入缓存，有效期24小时
	database.GlobalCache.Set(cacheKey, rates, 24*time.Hour)

	c.JSON(http.StatusOK, rates)
}

// ClearRatesCache 清除价格倍率缓存
func ClearRatesCache() {
	database.GlobalCache.Delete("price_rates")
}

// round 四舍五入到指定小数位
func round(num float64, precision int) float64 {
	precision10 := float64(1)
	for i := 0; i < precision; i++ {
		precision10 *= 10
	}
	return float64(int(num*precision10+0.5)) / precision10
}
