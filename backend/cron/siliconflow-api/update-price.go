package siliconflow_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"aimodels-prices/database"
	"aimodels-prices/handlers"
	"aimodels-prices/handlers/rates"
	"aimodels-prices/models"
)

// 常量定义
const (
	SiliconFlowChannelType = 45 // SiliconFlow的厂商ID
	SiliconFlowAPIEndpoint = "/api/v1/playground/comprehensive/all"
	SiliconFlowAPIHost     = "busy-bear.siliconflow.cn"
	PriceSource            = "SiliconFlow API"
	Status                 = "approved" // 设置为approved状态
	CreatedBy              = "cron自动任务"
	Currency               = "CNY" // 使用人民币
)

// 计费类型常量
const (
	BillingTypeTokens = "tokens" // 基于token的计费方式
	BillingTypeTimes  = "times"  // 基于次数的计费方式
)

// 定义API响应结构
type SiliconFlowResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    struct {
		Models []SiliconFlowModel `json:"models"`
	} `json:"data"`
}

// 模型信息结构
type SiliconFlowModel struct {
	ModelId             string   `json:"modelId"`
	ModelName           string   `json:"modelName"`
	DisplayName         string   `json:"DisplayName"`
	Mf                  string   `json:"mf"`
	Desc                string   `json:"desc"`
	Tags                []string `json:"tags"`
	Icon                string   `json:"icon"`
	Size                int      `json:"size"`
	ContextLen          int      `json:"contextLen"`
	Price               string   `json:"price"`
	Currency            string   `json:"currency"`
	PriceUnit           string   `json:"priceUnit"`
	Status              string   `json:"status"`
	Type                string   `json:"type"`
	SubType             string   `json:"subType"`
	JsonModeSupport     bool     `json:"jsonModeSupport"`
	FunctionCallSupport bool     `json:"functionCallSupport"`
}

// UpdateSiliconFlowPrices 更新SiliconFlow模型价格
func UpdateSiliconFlowPrices() error {
	log.Println("开始更新SiliconFlow价格数据...")

	// 获取API数据
	modelData, err := fetchSiliconFlowData()
	if err != nil {
		return fmt.Errorf("获取SiliconFlow数据失败: %v", err)
	}

	// 获取数据库连接
	db := database.DB
	if db == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	// 处理每个模型的价格数据
	processedCount := 0
	skippedCount := 0

	// 创建一个集合用于跟踪已处理的模型，避免重复
	processedModels := make(map[string]bool)

	for _, model := range modelData {
		modelName := model.ModelName

		// 检查是否已处理过这个模型
		if processedModels[modelName] {
			log.Printf("跳过已处理的模型: %s", modelName)
			skippedCount++
			continue
		}

		// 标记此模型为已处理
		processedModels[modelName] = true

		// 解析价格
		modelPrice, err := strconv.ParseFloat(model.Price, 64)
		if err != nil {
			log.Printf("解析价格失败 %s: %v", modelName, err)
			skippedCount++
			continue
		}

		// 确定模型类型和价格
		var modelType string
		var billingType string
		var inputPrice, outputPrice float64

		// 根据模型类型和价格单位确定模型类型和价格计算方式
		switch {
		case isTokenBasedUnit(model.PriceUnit):
			// 基于Token的模型（如文本模型）
			modelType = determineModelTypeBySubType(model.Type, model.SubType)
			billingType = BillingTypeTokens // 使用tokens计费类型
			// 直接使用价格，系统已经按每百万token为单位
			inputPrice = roundPrice(modelPrice)
			outputPrice = inputPrice // 使用相同价格
		case isTimeBasedUnit(model.PriceUnit, model.Type):
			// 基于次数的模型（如图像、视频）
			modelType = determineModelTypeBySubType(model.Type, model.SubType)
			billingType = BillingTypeTimes // 使用times计费类型
			// 直接使用价格
			inputPrice = roundPrice(modelPrice)
			outputPrice = inputPrice // 使用相同价格
		default:
			// 默认按token计费
			modelType = determineModelTypeBySubType(model.Type, model.SubType)
			// 根据模型类型决定计费方式
			if modelType == "text2image" || modelType == "text2video" || modelType == "image2video" {
				billingType = BillingTypeTimes // 图像和视频相关模型使用times
			} else {
				billingType = BillingTypeTokens // 其他默认使用tokens
			}
			// 对于未知类型，默认按token处理
			inputPrice = roundPrice(modelPrice)
			outputPrice = inputPrice // 使用相同价格
			log.Printf("未识别的价格单位: %s，默认使用计费类型: %s", model.PriceUnit, billingType)
		}

		// 创建价格对象
		price := models.Price{
			Model:       modelName,
			ModelType:   modelType,
			BillingType: billingType, // 使用动态确定的计费类型
			ChannelType: SiliconFlowChannelType,
			Currency:    Currency, // 使用人民币
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			PriceSource: PriceSource,
			Status:      Status, // 使用approved状态
			CreatedBy:   CreatedBy,
		}

		// 检查是否已存在相同模型的价格记录
		var existingPrice models.Price
		result := db.Where("model = ? AND channel_type = ?", modelName, SiliconFlowChannelType).First(&existingPrice)

		if result.Error == nil {
			// 使用processPrice函数处理更新，第三个参数设置为true表示直接审核通过
			_, changed, err := handlers.ProcessPrice(price, &existingPrice, true, CreatedBy)
			if err != nil {
				log.Printf("更新价格记录失败 %s: %v", modelName, err)
				skippedCount++
				continue
			}

			if changed {
				log.Printf("更新价格记录: %s", modelName)
				processedCount++
			} else {
				log.Printf("价格无变化，跳过更新: %s", modelName)
				skippedCount++
			}
		} else {
			// 检查是否存在相同模型名称的待审核记录
			var pendingCount int64
			if err := db.Model(&models.Price{}).Where("model = ? AND channel_type = ? AND status = 'pending'",
				modelName, SiliconFlowChannelType).Count(&pendingCount).Error; err != nil {
				log.Printf("检查待审核记录失败 %s: %v", modelName, err)
			}

			if pendingCount > 0 {
				log.Printf("已存在待审核的相同模型记录，跳过创建: %s", modelName)
				skippedCount++
				continue
			}

			// 使用processPrice函数处理创建，第三个参数设置为true表示直接审核通过
			_, changed, err := handlers.ProcessPrice(price, nil, true, CreatedBy)
			if err != nil {
				log.Printf("创建价格记录失败 %s: %v", modelName, err)
				skippedCount++
				continue
			}

			if changed {
				log.Printf("创建新价格记录: %s", modelName)
				processedCount++
			} else {
				log.Printf("价格创建失败: %s", modelName)
				skippedCount++
			}
		}
	}

	log.Printf("SiliconFlow价格数据处理完成，成功处理: %d, 跳过: %d", processedCount, skippedCount)

	// 清除倍率缓存
	rates.ClearRatesCache()
	log.Println("倍率缓存已清除")
	return nil
}

// roundPrice 对价格进行四舍五入处理，保留6位小数
func roundPrice(price float64) float64 {
	// 保留6位小数
	return math.Round(price*1000000) / 1000000
}

// fetchSiliconFlowData 获取SiliconFlow API数据
func fetchSiliconFlowData() ([]SiliconFlowModel, error) {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("环境变量SILICONFLOW_API_KEY未设置")
	}

	// 创建HTTPS连接
	conn, err := http.NewRequest("GET", "https://"+SiliconFlowAPIHost+SiliconFlowAPIEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	conn.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(conn)
	if err != nil {
		return nil, fmt.Errorf("请求SiliconFlow API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析JSON数据
	var siliconFlowResp SiliconFlowResponse
	if err := json.Unmarshal(body, &siliconFlowResp); err != nil {
		return nil, fmt.Errorf("解析JSON数据失败: %v", err)
	}

	// 检查响应状态
	if !siliconFlowResp.Status || siliconFlowResp.Code != 20000 {
		return nil, fmt.Errorf("API请求返回错误: %s", siliconFlowResp.Message)
	}

	return siliconFlowResp.Data.Models, nil
}

// isTokenBasedUnit 判断是否是基于token的计费单位
func isTokenBasedUnit(unit string) bool {
	tokenUnits := []string{
		"/ M Tokens",
		"/ M UTF-8 bytes",
		"/ M px / Steps",
	}

	for _, tokenUnit := range tokenUnits {
		if strings.Contains(unit, tokenUnit) {
			return true
		}
	}
	return false
}

// isTimeBasedUnit 判断是否是基于次数的计费单位
func isTimeBasedUnit(unit string, modelType string) bool {
	timeUnits := []string{
		"/ Video",
		"/ Image",
		"",
	}

	// 如果模型类型是视频或图像，即使价格单位为空也按次数计费
	if modelType == "video" || modelType == "image" {
		return true
	}

	for _, timeUnit := range timeUnits {
		if strings.Contains(unit, timeUnit) {
			return true
		}
	}
	return false
}

// determineModelTypeBySubType 根据模型类型和子类型确定我们系统中的模型类型
func determineModelTypeBySubType(modelType string, subType string) string {
	switch modelType {
	case "text":
		return "text2text"
	case "image":
		return "text2image"
	case "video":
		if subType == "image-to-video" {
			return "image2video"
		}
		return "text2video"
	case "audio":
		return "text2speech"
	case "embedding":
		return "embedding"
	default:
		return "other"
	}
}
