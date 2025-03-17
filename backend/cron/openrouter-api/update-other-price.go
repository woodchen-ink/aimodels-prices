package openrouter_api

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"encoding/json"
	"strings"

	"aimodels-prices/database"
	"aimodels-prices/handlers"
	"aimodels-prices/models"
)

// 定义厂商ID映射
var authorToChannelType = map[string]uint{
	"openai":    1,
	"anthropic": 14,
	"qwen":      17,
	"google":    25,
	"x-ai":      1001,
}

// 定义黑名单列表
var blacklist = []string{
	"shap-e",
	"palm-2",
	"o3-mini-high",
	"claude-instant",
	"claude-1",
	"claude-3-haiku",
	"claude-3-opus",
	"claude-3-sonnet",
}

const (
	OtherPriceSource = "三方API"
	OtherStatus      = "pending"
)

// UpdateOtherPrices 更新其他厂商的价格
func UpdateOtherPrices() error {
	log.Println("开始更新其他厂商价格数据...")

	// 复用已有的API请求获取数据
	resp, err := fetchOpenRouterData()
	if err != nil {
		return fmt.Errorf("获取OpenRouter数据失败: %v", err)
	}

	// 获取数据库连接
	db := database.DB
	if db == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	// 处理每个模型的价格数据
	processedCount := 0
	skippedCount := 0
	for _, modelData := range resp.Data {
		// 提取模型名称（slug中/后面的部分）
		parts := strings.Split(modelData.Slug, "/")
		if len(parts) < 2 {
			log.Printf("跳过无效的模型名称: %s", modelData.Slug)
			skippedCount++
			continue
		}

		// 获取模型名称并去除":free"后缀
		modelName := parts[1]
		modelName = strings.Split(modelName, ":")[0]

		// 检查是否在黑名单中
		if isInBlacklist(modelName) {
			log.Printf("跳过黑名单模型: %s", modelName)
			skippedCount++
			continue
		}

		// 获取作者名称
		author := parts[0]

		// 检查是否支持的厂商
		channelType, ok := authorToChannelType[author]
		if !ok {
			log.Printf("跳过不支持的厂商: %s", author)
			skippedCount++
			continue
		}

		// 处理特殊模型名称
		if author == "google" {
			// 处理gemini-flash-1.5系列模型名称
			if strings.HasPrefix(modelName, "gemini-flash-1.5") {
				suffix := strings.TrimPrefix(modelName, "gemini-flash-1.5")
				modelName = "gemini-1.5-flash" + suffix
				log.Printf("修正Google模型名称: %s -> %s", parts[1], modelName)
			}
		}
		if author == "anthropic" {
			// 处理claude-3.5-sonnet系列模型名称
			if strings.HasPrefix(modelName, "claude-3.5") {
				suffix := strings.TrimPrefix(modelName, "claude-3.5")
				modelName = "claude-3-5" + suffix
				log.Printf("修正Claude模型名称: %s -> %s", parts[1], modelName)
			}

			if strings.HasPrefix(modelName, "claude-3.7") {
				suffix := strings.TrimPrefix(modelName, "claude-3.7")
				modelName = "claude-3-7" + suffix
				log.Printf("修正Claude模型名称: %s -> %s", parts[1], modelName)
			}
		}

		// 确定模型类型
		modelType := determineModelType(modelData.Modality)

		// 解析价格
		var inputPrice, outputPrice float64
		var parseErr error

		// 如果输入或输出价格为空，直接跳过
		if modelData.Endpoint.Pricing.Prompt == "" || modelData.Endpoint.Pricing.Completion == "" {
			log.Printf("跳过价格数据不完整的模型: %s", modelData.Slug)
			skippedCount++
			continue
		}

		// 使用endpoint中的pricing
		if modelData.Endpoint.Pricing.Prompt != "" {
			inputPrice, parseErr = parsePrice(modelData.Endpoint.Pricing.Prompt)
			if parseErr != nil {
				log.Printf("解析endpoint输入价格失败 %s: %v", modelData.Slug, parseErr)
				skippedCount++
				continue
			}
		}

		if modelData.Endpoint.Pricing.Completion != "" {
			outputPrice, parseErr = parsePrice(modelData.Endpoint.Pricing.Completion)
			if parseErr != nil {
				log.Printf("解析endpoint输出价格失败 %s: %v", modelData.Slug, parseErr)
				skippedCount++
				continue
			}
		}

		// 创建价格对象
		price := models.Price{
			Model:       modelName,
			ModelType:   modelType,
			BillingType: BillingType,
			ChannelType: channelType,
			Currency:    Currency,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			PriceSource: OtherPriceSource,
			Status:      OtherStatus,
			CreatedBy:   CreatedBy,
		}

		// 检查是否已存在相同模型的价格记录
		var existingPrice models.Price
		result := db.Where("model = ? AND channel_type = ?", modelName, channelType).First(&existingPrice)

		if result.Error == nil {
			// 使用processPrice函数处理更新
			_, changed, err := handlers.ProcessPrice(price, &existingPrice, false, CreatedBy)
			if err != nil {
				log.Printf("更新价格记录失败 %s: %v", modelName, err)
				skippedCount++
				continue
			}

			if changed {
				log.Printf("更新价格记录: %s (厂商: %s)", modelName, author)
				processedCount++
			} else {
				log.Printf("价格无变化，跳过更新: %s (厂商: %s)", modelName, author)
				skippedCount++
			}
		} else {
			// 使用processPrice函数处理创建
			_, changed, err := handlers.ProcessPrice(price, nil, false, CreatedBy)
			if err != nil {
				log.Printf("创建价格记录失败 %s: %v", modelName, err)
				skippedCount++
				continue
			}

			if changed {
				log.Printf("创建新价格记录: %s (厂商: %s)", modelName, author)
				processedCount++
			} else {
				log.Printf("价格创建失败: %s (厂商: %s)", modelName, author)
				skippedCount++
			}
		}
	}

	log.Printf("其他厂商价格数据处理完成，成功处理: %d, 跳过: %d", processedCount, skippedCount)
	return nil
}

// fetchOpenRouterData 获取OpenRouter API数据
func fetchOpenRouterData() (*OpenRouterResponse, error) {
	// 复用已有的HTTP请求逻辑
	resp, err := http.Get(OpenRouterAPIURL)
	if err != nil {
		return nil, fmt.Errorf("请求OpenRouter API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析JSON数据
	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("解析JSON数据失败: %v", err)
	}

	return &openRouterResp, nil
}

// isInBlacklist 检查模型名称是否在黑名单中
func isInBlacklist(modelName string) bool {
	modelNameLower := strings.ToLower(modelName)
	for _, blacklistItem := range blacklist {
		if strings.Contains(modelNameLower, blacklistItem) {
			return true
		}
	}
	return false
}
