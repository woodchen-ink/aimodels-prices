package openai_api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"aimodels-prices/database"
	"aimodels-prices/handlers"
	"aimodels-prices/handlers/one_hub"
	"aimodels-prices/models"

	"golang.org/x/net/html"
	"gorm.io/gorm"
)

const (
	OpenAIPricingURL   = "https://developers.openai.com/api/docs/pricing"
	OpenAIChannelType  = 1
	BillingType        = "tokens"
	Currency           = "USD"
	PriceSource        = "https://developers.openai.com/api/docs/pricing"
	Status             = "approved"
	CreatedBy          = "cron自动任务"
)

// OpenAIModelPrice 从页面解析出的模型价格数据
type OpenAIModelPrice struct {
	Model       string
	InputPrice  float64 // $/1M tokens
	CachedPrice float64 // $/1M tokens
	OutputPrice float64 // $/1M tokens
}

// FetchAndSavePrices 抓取OpenAI官网价格页面并保存到数据库
func FetchAndSavePrices() error {
	log.Println("开始获取OpenAI官网价格数据...")

	// 抓取页面
	prices, err := fetchOpenAIPrices()
	if err != nil {
		return fmt.Errorf("获取OpenAI价格数据失败: %v", err)
	}

	if len(prices) == 0 {
		return fmt.Errorf("未解析到任何OpenAI价格数据")
	}

	log.Printf("成功解析到 %d 个OpenAI模型价格", len(prices))

	// 获取数据库连接
	db := database.DB
	if db == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	processedCount := 0
	skippedCount := 0

	for _, mp := range prices {
		// 构建Price对象
		price := models.Price{
			Model:       mp.Model,
			BillingType: BillingType,
			ChannelType: OpenAIChannelType,
			Currency:    Currency,
			InputPrice:  mp.InputPrice,
			OutputPrice: mp.OutputPrice,
			PriceSource: PriceSource,
			Status:      Status,
			CreatedBy:   CreatedBy,
		}

		// 设置缓存价格（如果有）
		if mp.CachedPrice > 0 {
			cachedPrice := mp.CachedPrice
			price.CachedTokens = &cachedPrice
		}

		// 检查是否已存在相同模型的价格记录
		var existingPrice models.Price
		result := db.Where("model = ? AND channel_type = ?", mp.Model, OpenAIChannelType).First(&existingPrice)

		if result.Error == nil {
			// 记录存在，执行更新
			_, changed, err := handlers.ProcessPrice(price, &existingPrice, true, CreatedBy)
			if err != nil {
				log.Printf("更新价格记录失败 %s: %v", mp.Model, err)
				skippedCount++
				continue
			}
			if changed {
				processedCount++
			} else {
				skippedCount++
			}
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录
			var pendingCount int64
			if err := db.Model(&models.Price{}).Where("model = ? AND channel_type = ? AND status = 'pending'",
				mp.Model, OpenAIChannelType).Count(&pendingCount).Error; err != nil {
				log.Printf("检查待审核记录失败 %s: %v", mp.Model, err)
			}

			if pendingCount > 0 {
				log.Printf("已存在待审核的相同模型记录，跳过创建: %s", mp.Model)
				skippedCount++
				continue
			}

			_, changed, err := handlers.ProcessPrice(price, nil, true, CreatedBy)
			if err != nil {
				log.Printf("创建价格记录失败 %s: %v", mp.Model, err)
				skippedCount++
				continue
			}
			if changed {
				processedCount++
			} else {
				log.Printf("价格创建失败: %s", mp.Model)
				skippedCount++
			}
		} else {
			log.Printf("查询价格记录时发生错误 %s: %v", mp.Model, result.Error)
			skippedCount++
		}
	}

	log.Printf("OpenAI官网价格数据处理完成，成功处理: %d, 跳过: %d", processedCount, skippedCount)

	// 清除倍率缓存
	one_hub.ClearRatesCache()
	log.Println("倍率缓存已清除")
	return nil
}

// fetchOpenAIPrices 抓取OpenAI定价页面并解析价格表格
func fetchOpenAIPrices() ([]OpenAIModelPrice, error) {
	req, err := http.NewRequest("GET", OpenAIPricingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PriceBot/1.0)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求OpenAI定价页面失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI定价页面返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	return parseHTMLPrices(string(body))
}

// parseHTMLPrices 从HTML中解析价格表格
// 只提取 Standard tier 的 Text tokens 表格（data-value="standard" 且列头为 Model/Input/Cached input/Output 的第一个表格）
func parseHTMLPrices(htmlContent string) ([]OpenAIModelPrice, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 查找 data-content-switcher-pane + data-value="standard" 的容器
	standardPanes := findStandardPanes(doc)
	if len(standardPanes) == 0 {
		return nil, fmt.Errorf("未找到 Standard tier 的价格面板")
	}

	// 在 standard pane 中查找第一个4列（Model/Input/Cached input/Output）的表格
	for _, pane := range standardPanes {
		tables := findElements(pane, "table")
		for _, table := range tables {
			headers := getTableHeaders(table)
			if len(headers) != 4 {
				continue
			}
			// 确认是 Model/Input/Cached*/Output 结构（排除带 Training 列的 Fine-tuning 表格）
			h0 := strings.ToLower(strings.TrimSpace(headers[0]))
			h1 := strings.ToLower(strings.TrimSpace(headers[1]))
			h3 := strings.ToLower(strings.TrimSpace(headers[3]))
			if h0 == "model" && h1 == "input" && h3 == "output" {
				prices := parseTable(table)
				if len(prices) > 0 {
					return prices, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("未找到 Standard tier 的 Text tokens 价格表格")
}

// findStandardPanes 查找所有 data-content-switcher-pane + data-value="standard" 的元素
func findStandardPanes(n *html.Node) []*html.Node {
	var result []*html.Node
	if n.Type == html.ElementNode {
		isPane := false
		isStandard := false
		for _, attr := range n.Attr {
			if attr.Key == "data-content-switcher-pane" {
				isPane = true
			}
			if attr.Key == "data-value" && attr.Val == "standard" {
				isStandard = true
			}
		}
		if isPane && isStandard {
			result = append(result, n)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findStandardPanes(c)...)
	}
	return result
}

// parseTable 解析单个表格，提取价格数据
func parseTable(table *html.Node) []OpenAIModelPrice {
	// 获取thead中的列头
	headers := getTableHeaders(table)
	if len(headers) == 0 {
		return nil
	}

	// 检查是否为价格表格（需要有Model、Input、Output列）
	colMap := make(map[string]int)
	for i, h := range headers {
		hLower := strings.ToLower(strings.TrimSpace(h))
		switch {
		case hLower == "model":
			colMap["model"] = i
		case hLower == "input":
			colMap["input"] = i
		case strings.Contains(hLower, "cached"):
			colMap["cached"] = i
		case hLower == "output":
			colMap["output"] = i
		}
	}

	// 必须同时有model、input、output列
	if _, ok := colMap["model"]; !ok {
		return nil
	}
	if _, ok := colMap["input"]; !ok {
		return nil
	}
	if _, ok := colMap["output"]; !ok {
		return nil
	}

	// 获取tbody中的数据行
	rows := getTableRows(table)
	var prices []OpenAIModelPrice

	for _, row := range rows {
		cells := getRowCells(row)
		if len(cells) <= colMap["output"] {
			continue
		}

		modelName := strings.TrimSpace(cells[colMap["model"]])
		if modelName == "" {
			continue
		}

		inputStr := strings.TrimSpace(cells[colMap["input"]])
		outputStr := strings.TrimSpace(cells[colMap["output"]])

		inputPrice, err := parseDollarPrice(inputStr)
		if err != nil {
			log.Printf("解析输入价格失败 %s: %v", modelName, err)
			continue
		}

		outputPrice, err := parseDollarPrice(outputStr)
		if err != nil {
			log.Printf("解析输出价格失败 %s: %v", modelName, err)
			continue
		}

		mp := OpenAIModelPrice{
			Model:       modelName,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
		}

		// 解析缓存价格（可选）
		if cachedIdx, ok := colMap["cached"]; ok && cachedIdx < len(cells) {
			cachedStr := strings.TrimSpace(cells[cachedIdx])
			cachedPrice, err := parseDollarPrice(cachedStr)
			if err == nil {
				mp.CachedPrice = cachedPrice
			}
		}

		prices = append(prices, mp)
	}

	return prices
}

// parseDollarPrice 解析美元价格字符串，如 "$2.50" -> 2.5 (per 1M tokens)
// 页面价格已经是 per 1M tokens 为单位
func parseDollarPrice(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || s == "—" {
		return 0, nil
	}

	// 移除 $ 符号
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimSpace(s)

	price, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析价格: %s", s)
	}

	// 四舍五入到6位小数
	result := math.Round(price*1000000) / 1000000
	return result, nil
}

// findElements 递归查找指定标签名的所有元素
func findElements(n *html.Node, tag string) []*html.Node {
	var result []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		result = append(result, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findElements(c, tag)...)
	}
	return result
}

// getTextContent 递归获取节点的文本内容
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

// getTableHeaders 获取表格thead中的列头文本
func getTableHeaders(table *html.Node) []string {
	var headers []string
	// 先找thead
	theads := findElements(table, "thead")
	if len(theads) == 0 {
		return nil
	}
	// 在thead中找th
	ths := findElements(theads[0], "th")
	for _, th := range ths {
		headers = append(headers, getTextContent(th))
	}
	return headers
}

// getTableRows 获取表格tbody中的所有tr
func getTableRows(table *html.Node) []*html.Node {
	tbodies := findElements(table, "tbody")
	if len(tbodies) == 0 {
		return nil
	}
	return findElements(tbodies[0], "tr")
}

// getRowCells 获取一行中所有td的文本
func getRowCells(tr *html.Node) []string {
	var cells []string
	tds := findElements(tr, "td")
	for _, td := range tds {
		cells = append(cells, getTextContent(td))
	}
	return cells
}
