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
	":extended",
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

	// 创建一个映射，用于按作者和模型名称存储模型数据
	// 键：作者/模型名称基础部分
	// 值：带有free标识和不带free标识的模型数据
	modelDataMap := make(map[string]map[bool]*ModelData)

	// 第一遍遍历，分类整理模型数据
	for _, modelData := range resp.Data {
		// 提取模型名称（slug中/后面的部分）
		parts := strings.Split(modelData.Slug, "/")
		if len(parts) < 2 {
			log.Printf("跳过无效的模型名称: %s", modelData.Slug)
			continue
		}

		author := parts[0]
		fullModelName := parts[1]

		// 判断是否带有":free"后缀
		isFree := strings.HasSuffix(fullModelName, ":free")

		// 提取基础模型名称（不带":free"后缀）
		baseModelName := fullModelName
		if isFree {
			baseModelName = strings.TrimSuffix(fullModelName, ":free")
		}

		// 创建模型的唯一键
		modelKey := author + "/" + baseModelName

		// 如果需要，为这个模型键初始化一个条目
		if _, exists := modelDataMap[modelKey]; !exists {
			modelDataMap[modelKey] = make(map[bool]*ModelData)
		}

		// 存储模型数据
		modelDataMap[modelKey][isFree] = &modelData
	}

	// 第二遍遍历，根据处理规则选择合适的模型数据
	for modelKey, variants := range modelDataMap {
		var modelData *ModelData

		// 优先选择非free版本
		if nonFreeData, hasNonFree := variants[false]; hasNonFree {
			modelData = nonFreeData
		} else if freeData, hasFree := variants[true]; hasFree {
			// 如果只有free版本，则使用free版本
			modelData = freeData
		} else {
			// 不应该发生，但为了安全
			log.Printf("处理模型数据异常: %s", modelKey)
			skippedCount++
			continue
		}

		// 提取模型名称
		parts := strings.Split(modelData.Slug, "/")
		modelName := strings.Split(parts[1], ":")[0] // 移除":free"后缀
		author := parts[0]

		// 检查是否在黑名单中
		if isInBlacklist(modelName) {
			log.Printf("跳过黑名单模型: %s", modelName)
			skippedCount++
			continue
		}

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
