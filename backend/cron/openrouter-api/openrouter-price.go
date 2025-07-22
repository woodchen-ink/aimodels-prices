package openrouter_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"aimodels-prices/database"
	"aimodels-prices/handlers"
	"aimodels-prices/models"
)

const (
	OpenRouterAPIURL = "https://openrouter.ai/api/frontend/models"
	ChannelType      = 20
	BillingType      = "tokens"
	Currency         = "USD"
	PriceSource      = "https://openrouter.ai/models"
	Status           = "approved"
	CreatedBy        = "cron自动任务"
)

type OpenRouterResponse struct {
	Data []ModelData `json:"data"`
}

type ModelData struct {
	Slug     string   `json:"slug"`
	Modality string   `json:"modality"`
	Pricing  Pricing  `json:"pricing"`
	Endpoint Endpoint `json:"endpoint"`
}

type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

type Endpoint struct {
	Pricing Pricing `json:"pricing"`
}

// FetchAndSavePrices 获取OpenRouter API的价格并保存到数据库
func FetchAndSavePrices() error {
	log.Println("开始获取OpenRouter价格数据...")

	// 发送GET请求获取数据
	resp, err := http.Get(OpenRouterAPIURL)
	if err != nil {
		return fmt.Errorf("请求OpenRouter API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析JSON数据
	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return fmt.Errorf("解析JSON数据失败: %v", err)
	}

	// 获取数据库连接
	db := database.DB
	if db == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	// 处理每个模型的价格数据
	processedCount := 0
	skippedCount := 0
	for _, modelData := range openRouterResp.Data {
		// 1. 检查API返回的模型是否有价格字段，如果没有价格则跳过
		hasPricing := false
		if (modelData.Endpoint.Pricing.Prompt != "" && modelData.Endpoint.Pricing.Completion != "") ||
			(modelData.Pricing.Prompt != "" && modelData.Pricing.Completion != "") {
			hasPricing = true
		}

		if !hasPricing {
			log.Printf("跳过无价格模型: %s", modelData.Slug)
			skippedCount++
			continue
		}

		// 2. 检查模型名称是否包含":free"，如果是免费模型则设置价格为0
		isFreeModel := strings.Contains(modelData.Slug, ":free")
		
		// 确定模型类型
		modelType := determineModelType(modelData.Modality)

		// 使用endpoint中的pricing
		var inputPrice, outputPrice float64
		var err error

		if isFreeModel {
			// 免费模型价格设置为0
			inputPrice = 0
			outputPrice = 0
			log.Printf("处理免费模型，价格设为0: %s", modelData.Slug)
		} else {
			// 优先使用endpoint中的pricing
			if modelData.Endpoint.Pricing.Prompt != "" {
				inputPrice, err = parsePrice(modelData.Endpoint.Pricing.Prompt)
				if err != nil {
					log.Printf("解析endpoint输入价格失败 %s: %v", modelData.Slug, err)
					skippedCount++
					continue
				}
			} else if modelData.Pricing.Prompt != "" {
				// 如果endpoint中没有，则使用顶层pricing
				inputPrice, err = parsePrice(modelData.Pricing.Prompt)
				if err != nil {
					log.Printf("解析输入价格失败 %s: %v", modelData.Slug, err)
					skippedCount++
					continue
				}
			}

			if modelData.Endpoint.Pricing.Completion != "" {
				outputPrice, err = parsePrice(modelData.Endpoint.Pricing.Completion)
				if err != nil {
					log.Printf("解析endpoint输出价格失败 %s: %v", modelData.Slug, err)
					skippedCount++
					continue
				}
			} else if modelData.Pricing.Completion != "" {
				outputPrice, err = parsePrice(modelData.Pricing.Completion)
				if err != nil {
					log.Printf("解析输出价格失败 %s: %v", modelData.Slug, err)
					skippedCount++
					continue
				}
			}
		}

		// 创建价格对象
		price := models.Price{
			Model:       modelData.Slug,
			ModelType:   modelType,
			BillingType: BillingType,
			ChannelType: ChannelType,
			Currency:    Currency,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			PriceSource: PriceSource,
			Status:      Status,
			CreatedBy:   CreatedBy,
		}

		// 检查是否已存在相同模型的价格记录
		var existingPrice models.Price
		result := db.Where("model = ? AND channel_type = ?", modelData.Slug, ChannelType).First(&existingPrice)

		if result.Error == nil {
			// 使用processPrice函数处理更新
			_, changed, err := handlers.ProcessPrice(price, &existingPrice, true, CreatedBy)
			if err != nil {
				log.Printf("更新价格记录失败 %s: %v", modelData.Slug, err)
				skippedCount++
				continue
			}

			if changed {
				log.Printf("更新价格记录: %s", modelData.Slug)
				processedCount++
			} else {
				// log.Printf("价格无变化，跳过更新: %s", modelData.Slug)
				skippedCount++
			}
		} else {
			// 使用processPrice函数处理创建
			_, changed, err := handlers.ProcessPrice(price, nil, true, CreatedBy)
			if err != nil {
				log.Printf("创建价格记录失败 %s: %v", modelData.Slug, err)
				skippedCount++
				continue
			}

			if changed {
				// log.Printf("创建新价格记录: %s", modelData.Slug)
				processedCount++
			} else {
				log.Printf("价格创建失败: %s", modelData.Slug)
				skippedCount++
			}
		}
	}

	log.Printf("OpenRouter价格数据处理完成，成功处理: %d, 跳过: %d", processedCount, skippedCount)
	return nil
}

// determineModelType 根据modality确定模型类型
func determineModelType(modality string) string {
	switch modality {
	case "text->text":
		return "text2text"
	case "text+image->text":
		return "multimodal"
	default:
		return "other"
	}
}

// parsePrice 解析价格字符串为浮点数并乘以1000000
func parsePrice(priceStr string) (float64, error) {
	if priceStr == "" {
		return 0, nil // 如果价格为空，返回0
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Printf("价格解析失败: %s, 错误: %v", priceStr, err)
		return 0, err
	}

	// 乘以1000000并四舍五入到6位小数，避免浮点数精度问题
	result := math.Round(price*1000000*1000000) / 1000000
	return result, nil
}
