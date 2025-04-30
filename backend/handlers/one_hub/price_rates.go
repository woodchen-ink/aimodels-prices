package one_hub

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// ExtraRatios 扩展价格倍率结构
type ExtraRatios struct {
	InputAudioTokens  *float64 `json:"input_audio_tokens,omitempty"`
	OutputAudioTokens *float64 `json:"output_audio_tokens,omitempty"`
	CachedTokens      *float64 `json:"cached_tokens,omitempty"`
	CachedReadTokens  *float64 `json:"cached_read_tokens,omitempty"`
	CachedWriteTokens *float64 `json:"cached_write_tokens,omitempty"`
	ReasoningTokens   *float64 `json:"reasoning_tokens,omitempty"`
	InputTextTokens   *float64 `json:"input_text_tokens,omitempty"`
	OutputTextTokens  *float64 `json:"output_text_tokens,omitempty"`
	InputImageTokens  *float64 `json:"input_image_tokens,omitempty"`
	OutputImageTokens *float64 `json:"output_image_tokens,omitempty"`
}

// PriceRate 价格倍率结构
type PriceRate struct {
	Model       string       `json:"model"`
	Type        string       `json:"type"`
	ChannelType uint         `json:"channel_type"`
	Input       float64      `json:"input"`
	Output      float64      `json:"output"`
	ExtraRatios *ExtraRatios `json:"extra_ratios,omitempty"`
}

// GetPriceRates 获取价格倍率
func GetPriceRates(c *gin.Context) {
	cacheKey := "one_hub_price_rates"

	// 尝试从缓存获取
	if cachedData, found := database.GlobalCache.Get(cacheKey); found {
		if rates, ok := cachedData.([]PriceRate); ok {
			c.JSON(http.StatusOK, rates)
			return
		}
	}

	// 使用索引优化查询，只查询需要的字段
	var prices []models.Price
	if err := database.DB.Select("model, billing_type, channel_type, input_price, output_price, currency, status, input_audio_tokens, output_audio_tokens, cached_tokens, cached_read_tokens, cached_write_tokens, reasoning_tokens, input_text_tokens, output_text_tokens, input_image_tokens, output_image_tokens").
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

		// 创建额外价格倍率
		var extraRatios *ExtraRatios

		// 只有当至少有一个扩展价格字段不为nil时才创建ExtraRatios
		if price.InputAudioTokens != nil || price.OutputAudioTokens != nil ||
			price.CachedTokens != nil || price.CachedReadTokens != nil || price.CachedWriteTokens != nil ||
			price.ReasoningTokens != nil || price.InputTextTokens != nil || price.OutputTextTokens != nil ||
			price.InputImageTokens != nil || price.OutputImageTokens != nil {

			extraRatios = &ExtraRatios{}

			// 计算各扩展价格字段的倍率
			if price.InputAudioTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputAudioTokens/2, 4)
				} else {
					rate = round(*price.InputAudioTokens/14, 4)
				}
				extraRatios.InputAudioTokens = &rate
			}

			if price.OutputAudioTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputAudioTokens/2, 4)
				} else {
					rate = round(*price.OutputAudioTokens/14, 4)
				}
				extraRatios.OutputAudioTokens = &rate
			}

			if price.CachedTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedTokens/2, 4)
				} else {
					rate = round(*price.CachedTokens/14, 4)
				}
				extraRatios.CachedTokens = &rate
			}

			if price.CachedReadTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedReadTokens/2, 4)
				} else {
					rate = round(*price.CachedReadTokens/14, 4)
				}
				extraRatios.CachedReadTokens = &rate
			}

			if price.CachedWriteTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedWriteTokens/2, 4)
				} else {
					rate = round(*price.CachedWriteTokens/14, 4)
				}
				extraRatios.CachedWriteTokens = &rate
			}

			if price.ReasoningTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.ReasoningTokens/2, 4)
				} else {
					rate = round(*price.ReasoningTokens/14, 4)
				}
				extraRatios.ReasoningTokens = &rate
			}

			if price.InputTextTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputTextTokens/2, 4)
				} else {
					rate = round(*price.InputTextTokens/14, 4)
				}
				extraRatios.InputTextTokens = &rate
			}

			if price.OutputTextTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputTextTokens/2, 4)
				} else {
					rate = round(*price.OutputTextTokens/14, 4)
				}
				extraRatios.OutputTextTokens = &rate
			}

			if price.InputImageTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputImageTokens/2, 4)
				} else {
					rate = round(*price.InputImageTokens/14, 4)
				}
				extraRatios.InputImageTokens = &rate
			}

			if price.OutputImageTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputImageTokens/2, 4)
				} else {
					rate = round(*price.OutputImageTokens/14, 4)
				}
				extraRatios.OutputImageTokens = &rate
			}
		}

		// 创建当前价格的PriceRate
		currentRate := PriceRate{
			Model:       price.Model,
			Type:        price.BillingType,
			ChannelType: price.ChannelType,
			Input:       inputRate,
			Output:      outputRate,
			ExtraRatios: extraRatios,
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

// GetOfficialPriceRates 获取官方厂商（ID小于1000）的价格倍率
func GetOfficialPriceRates(c *gin.Context) {
	cacheKey := "one_hub_official_price_rates"

	// 尝试从缓存获取
	if cachedData, found := database.GlobalCache.Get(cacheKey); found {
		if rates, ok := cachedData.([]PriceRate); ok {
			c.JSON(http.StatusOK, rates)
			return
		}
	}

	// 使用索引优化查询，只查询需要的字段，并添加厂商ID筛选条件
	var prices []models.Price
	result := database.DB.Model(&models.Price{}).
		Select("model, billing_type, channel_type, input_price, output_price, currency, status, input_audio_tokens, output_audio_tokens, cached_tokens, cached_read_tokens, cached_write_tokens, reasoning_tokens, input_text_tokens, output_text_tokens, input_image_tokens, output_image_tokens").
		Where(&models.Price{Status: "approved"}).
		Where("channel_type < ?", 1000).
		Find(&prices)

	if result.Error != nil {
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

		// 创建额外价格倍率
		var extraRatios *ExtraRatios

		// 只有当至少有一个扩展价格字段不为nil时才创建ExtraRatios
		if price.InputAudioTokens != nil || price.OutputAudioTokens != nil ||
			price.CachedTokens != nil || price.CachedReadTokens != nil || price.CachedWriteTokens != nil ||
			price.ReasoningTokens != nil || price.InputTextTokens != nil || price.OutputTextTokens != nil ||
			price.InputImageTokens != nil || price.OutputImageTokens != nil {

			extraRatios = &ExtraRatios{}

			// 计算各扩展价格字段的倍率
			if price.InputAudioTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputAudioTokens/2, 4)
				} else {
					rate = round(*price.InputAudioTokens/14, 4)
				}
				extraRatios.InputAudioTokens = &rate
			}

			if price.OutputAudioTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputAudioTokens/2, 4)
				} else {
					rate = round(*price.OutputAudioTokens/14, 4)
				}
				extraRatios.OutputAudioTokens = &rate
			}

			if price.CachedTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedTokens/2, 4)
				} else {
					rate = round(*price.CachedTokens/14, 4)
				}
				extraRatios.CachedTokens = &rate
			}

			if price.CachedReadTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedReadTokens/2, 4)
				} else {
					rate = round(*price.CachedReadTokens/14, 4)
				}
				extraRatios.CachedReadTokens = &rate
			}

			if price.CachedWriteTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.CachedWriteTokens/2, 4)
				} else {
					rate = round(*price.CachedWriteTokens/14, 4)
				}
				extraRatios.CachedWriteTokens = &rate
			}

			if price.ReasoningTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.ReasoningTokens/2, 4)
				} else {
					rate = round(*price.ReasoningTokens/14, 4)
				}
				extraRatios.ReasoningTokens = &rate
			}

			if price.InputTextTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputTextTokens/2, 4)
				} else {
					rate = round(*price.InputTextTokens/14, 4)
				}
				extraRatios.InputTextTokens = &rate
			}

			if price.OutputTextTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputTextTokens/2, 4)
				} else {
					rate = round(*price.OutputTextTokens/14, 4)
				}
				extraRatios.OutputTextTokens = &rate
			}

			if price.InputImageTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.InputImageTokens/2, 4)
				} else {
					rate = round(*price.InputImageTokens/14, 4)
				}
				extraRatios.InputImageTokens = &rate
			}

			if price.OutputImageTokens != nil {
				var rate float64
				if price.Currency == "USD" {
					rate = round(*price.OutputImageTokens/2, 4)
				} else {
					rate = round(*price.OutputImageTokens/14, 4)
				}
				extraRatios.OutputImageTokens = &rate
			}
		}

		// 创建当前价格的PriceRate
		currentRate := PriceRate{
			Model:       price.Model,
			Type:        price.BillingType,
			ChannelType: price.ChannelType,
			Input:       inputRate,
			Output:      outputRate,
			ExtraRatios: extraRatios,
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
	database.GlobalCache.Delete("one_hub_price_rates")
	database.GlobalCache.Delete("one_hub_official_price_rates")
}

// round 四舍五入到指定小数位
func round(num float64, precision int) float64 {
	precision10 := float64(1)
	for i := 0; i < precision; i++ {
		precision10 *= 10
	}
	return float64(int(num*precision10+0.5)) / precision10
}
