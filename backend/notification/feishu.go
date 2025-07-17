package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"aimodels-prices/models"
)

// FeishuWebhook 飞书webhook配置
type FeishuWebhook struct {
	URL string
}

// TextMessage 文本消息结构
type TextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// CardMessage 卡片消息结构
type CardMessage struct {
	MsgType string `json:"msg_type"`
	Card    Card   `json:"card"`
}

// Card 卡片结构
type Card struct {
	Schema string     `json:"schema"`
	Config CardConfig `json:"config"`
	Header CardHeader `json:"header"`
	Body   CardBody   `json:"body"`
}

// CardConfig 卡片配置
type CardConfig struct {
	UpdateMulti bool `json:"update_multi"`
}

// CardHeader 卡片头部
type CardHeader struct {
	Title    Title  `json:"title"`
	Template string `json:"template"`
	Padding  string `json:"padding"`
}

// Title 标题
type Title struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// CardBody 卡片主体
type CardBody struct {
	Direction string        `json:"direction"`
	Padding   string        `json:"padding"`
	Elements  []CardElement `json:"elements"`
}

// CardElement 卡片元素
type CardElement struct {
	Tag       string `json:"tag"`
	Content   string `json:"content,omitempty"`
	TextAlign string `json:"text_align,omitempty"`
	TextSize  string `json:"text_size,omitempty"`
	Margin    string `json:"margin,omitempty"`
}

// NewFeishuWebhook 创建飞书webhook实例
func NewFeishuWebhook() *FeishuWebhook {
	url := os.Getenv("FEISHU_WEBHOOK_URL")
	if url == "" {
		return nil
	}
	return &FeishuWebhook{URL: url}
}

// SendTextMessage 发送文本消息
func (f *FeishuWebhook) SendTextMessage(text string) error {
	if f == nil || f.URL == "" {
		return nil // 如果没有配置webhook，则跳过
	}

	message := TextMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: text,
		},
	}

	return f.sendMessage(message)
}

// SendPendingPriceNotification 发送待审核价格通知卡片
func (f *FeishuWebhook) SendPendingPriceNotification(price models.Price, providerName string, isNew bool) error {
	if f == nil || f.URL == "" {
		return nil // 如果没有配置webhook，则跳过
	}

	var actionText string
	if isNew {
		actionText = "新增"
	} else {
		actionText = "更新"
	}

	// 构建卡片内容
	content := fmt.Sprintf("**%s模型价格**\n\n", actionText)
	content += fmt.Sprintf("**模型名称：** %s\n", getDisplayModel(price))
	content += fmt.Sprintf("**厂商：** %s\n", providerName)
	content += fmt.Sprintf("**计费类型：** %s\n", getBillingTypeText(getDisplayBillingType(price)))
	content += fmt.Sprintf("**输入价格：** %.6f %s/1K tokens\n", getDisplayInputPrice(price), getDisplayCurrency(price))
	content += fmt.Sprintf("**输出价格：** %.6f %s/1K tokens\n", getDisplayOutputPrice(price), getDisplayCurrency(price))
	content += fmt.Sprintf("**创建者：** %s\n", price.CreatedBy)
	content += fmt.Sprintf("**创建时间：** %s", time.Now().Format("2006-01-02 15:04:05"))

	// 如果有扩展价格字段，也显示出来
	if hasExtendedPrices(price) {
		content += "\n\n**扩展价格：**\n"
		if getDisplayInputAudioTokens(price) != nil {
			content += fmt.Sprintf("- 音频输入：%.6f %s/1K tokens\n", *getDisplayInputAudioTokens(price), getDisplayCurrency(price))
		}
		if getDisplayOutputAudioTokens(price) != nil {
			content += fmt.Sprintf("- 音频输出：%.6f %s/1K tokens\n", *getDisplayOutputAudioTokens(price), getDisplayCurrency(price))
		}
		if getDisplayCachedTokens(price) != nil {
			content += fmt.Sprintf("- 缓存：%.6f %s/1K tokens\n", *getDisplayCachedTokens(price), getDisplayCurrency(price))
		}
		if getDisplayReasoningTokens(price) != nil {
			content += fmt.Sprintf("- 推理：%.6f %s/1K tokens\n", *getDisplayReasoningTokens(price), getDisplayCurrency(price))
		}
	}

	card := CardMessage{
		MsgType: "interactive",
		Card: Card{
			Schema: "2.0",
			Config: CardConfig{
				UpdateMulti: true,
			},
			Header: CardHeader{
				Title: Title{
					Tag:     "plain_text",
					Content: fmt.Sprintf("🔔 有新的价格待审核 - %s", actionText),
				},
				Template: "orange",
				Padding:  "12px 12px 12px 12px",
			},
			Body: CardBody{
				Direction: "vertical",
				Padding:   "12px 12px 12px 12px",
				Elements: []CardElement{
					{
						Tag:       "markdown",
						Content:   content,
						TextAlign: "left",
						TextSize:  "normal",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}

	return f.sendMessage(card)
}

// SendBatchNotification 发送批量通知
func (f *FeishuWebhook) SendBatchNotification(count int) error {
	if f == nil || f.URL == "" {
		return nil
	}

	content := fmt.Sprintf("📢 **批量价格更新通知**\n\n本次共有 **%d** 个模型价格等待审核，请及时处理。", count)

	card := CardMessage{
		MsgType: "interactive",
		Card: Card{
			Schema: "2.0",
			Config: CardConfig{
				UpdateMulti: true,
			},
			Header: CardHeader{
				Title: Title{
					Tag:     "plain_text",
					Content: "📊 批量价格更新通知",
				},
				Template: "blue",
				Padding:  "12px 12px 12px 12px",
			},
			Body: CardBody{
				Direction: "vertical",
				Padding:   "12px 12px 12px 12px",
				Elements: []CardElement{
					{
						Tag:       "markdown",
						Content:   content,
						TextAlign: "left",
						TextSize:  "normal",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}

	return f.sendMessage(card)
}

// SendPendingPricesDetailedNotification 发送详细的待审核价格统计通知
func (f *FeishuWebhook) SendPendingPricesDetailedNotification(content string, count int) error {
	if f == nil || f.URL == "" {
		return nil
	}

	card := CardMessage{
		MsgType: "interactive",
		Card: Card{
			Schema: "2.0",
			Config: CardConfig{
				UpdateMulti: true,
			},
			Header: CardHeader{
				Title: Title{
					Tag:     "plain_text",
					Content: fmt.Sprintf("🔍 待审核价格检查报告 - %d个待审核", count),
				},
				Template: "red",
				Padding:  "12px 12px 12px 12px",
			},
			Body: CardBody{
				Direction: "vertical",
				Padding:   "12px 12px 12px 12px",
				Elements: []CardElement{
					{
						Tag:       "markdown",
						Content:   content,
						TextAlign: "left",
						TextSize:  "normal",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}

	return f.sendMessage(card)
}

// sendMessage 发送消息到飞书
func (f *FeishuWebhook) sendMessage(message interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	resp, err := http.Post(f.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

// 辅助函数：获取显示用的模型名称
func getDisplayModel(price models.Price) string {
	if price.TempModel != nil {
		return *price.TempModel
	}
	return price.Model
}

// 辅助函数：获取显示用的计费类型
func getDisplayBillingType(price models.Price) string {
	if price.TempBillingType != nil {
		return *price.TempBillingType
	}
	return price.BillingType
}

// 辅助函数：获取显示用的货币
func getDisplayCurrency(price models.Price) string {
	if price.TempCurrency != nil {
		return *price.TempCurrency
	}
	return price.Currency
}

// 辅助函数：获取显示用的输入价格
func getDisplayInputPrice(price models.Price) float64 {
	if price.TempInputPrice != nil {
		return *price.TempInputPrice
	}
	return price.InputPrice
}

// 辅助函数：获取显示用的输出价格
func getDisplayOutputPrice(price models.Price) float64 {
	if price.TempOutputPrice != nil {
		return *price.TempOutputPrice
	}
	return price.OutputPrice
}

// 辅助函数：获取显示用的音频输入价格
func getDisplayInputAudioTokens(price models.Price) *float64 {
	if price.TempInputAudioTokens != nil {
		return price.TempInputAudioTokens
	}
	return price.InputAudioTokens
}

// 辅助函数：获取显示用的音频输出价格
func getDisplayOutputAudioTokens(price models.Price) *float64 {
	if price.TempOutputAudioTokens != nil {
		return price.TempOutputAudioTokens
	}
	return price.OutputAudioTokens
}

// 辅助函数：获取显示用的缓存价格
func getDisplayCachedTokens(price models.Price) *float64 {
	if price.TempCachedTokens != nil {
		return price.TempCachedTokens
	}
	return price.CachedTokens
}

// 辅助函数：获取显示用的推理价格
func getDisplayReasoningTokens(price models.Price) *float64 {
	if price.TempReasoningTokens != nil {
		return price.TempReasoningTokens
	}
	return price.ReasoningTokens
}

// 辅助函数：检查是否有扩展价格字段
func hasExtendedPrices(price models.Price) bool {
	return getDisplayInputAudioTokens(price) != nil ||
		getDisplayOutputAudioTokens(price) != nil ||
		getDisplayCachedTokens(price) != nil ||
		getDisplayReasoningTokens(price) != nil
}

// 辅助函数：获取计费类型中文显示
func getBillingTypeText(billingType string) string {
	switch billingType {
	case "token":
		return "按Token计费"
	case "request":
		return "按请求计费"
	case "minute":
		return "按分钟计费"
	case "hour":
		return "按小时计费"
	case "day":
		return "按天计费"
	case "month":
		return "按月计费"
	default:
		return billingType
	}
}
