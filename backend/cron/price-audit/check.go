package price_audit

import (
	"fmt"
	"log"
	"time"

	"aimodels-prices/database"
	"aimodels-prices/models"
	"aimodels-prices/notification"
)

// lastNotificationTime 记录上次发送通知的时间，避免重复发送
var lastNotificationTime time.Time

// CheckPendingPrices 检查待审核价格并发送通知
func CheckPendingPrices() error {
	log.Println("开始检查待审核价格...")

	// 查询所有待审核的价格
	var pendingPrices []models.Price
	if err := database.DB.Where("status = 'pending'").Find(&pendingPrices).Error; err != nil {
		log.Printf("查询待审核价格失败: %v", err)
		return err
	}

	if len(pendingPrices) == 0 {
		log.Println("当前没有待审核的价格")
		return nil
	}

	log.Printf("发现 %d 个待审核价格", len(pendingPrices))

	// 检查是否需要发送通知（避免频繁发送）
	now := time.Now()
	if now.Sub(lastNotificationTime) < 24*time.Hour {
		log.Println("距离上次通知时间较短，跳过本次通知")
		return nil
	}

	// 发送飞书通知
	webhook := notification.NewFeishuWebhook()
	if webhook == nil {
		log.Println("未配置飞书webhook，跳过通知")
		return nil
	}

	// 异步发送通知
	go func() {
		if err := sendPendingPricesNotification(webhook, pendingPrices); err != nil {
			log.Printf("发送飞书通知失败: %v", err)
		} else {
			log.Printf("成功发送飞书通知，包含 %d 个待审核价格", len(pendingPrices))
			lastNotificationTime = now
		}
	}()

	return nil
}

// sendPendingPricesNotification 发送待审核价格的详细通知
func sendPendingPricesNotification(webhook *notification.FeishuWebhook, pendingPrices []models.Price) error {
	// 按厂商分组统计
	providerStats := make(map[uint][]models.Price)
	for _, price := range pendingPrices {
		channelType := getChannelType(price)
		providerStats[channelType] = append(providerStats[channelType], price)
	}

	// 构建详细的通知内容
	content := fmt.Sprintf("📋 **待审核价格统计**\n\n**总计：** %d 个模型价格待审核\n\n", len(pendingPrices))

	// 按厂商分组显示
	content += "**分厂商统计：**\n"
	for channelType, prices := range providerStats {
		var provider models.Provider
		if err := database.DB.Where("id = ?", channelType).First(&provider).Error; err != nil {
			provider.Name = fmt.Sprintf("厂商ID:%d", channelType)
		}
		content += fmt.Sprintf("- %s：%d 个模型\n", provider.Name, len(prices))
	}

	// 显示最近的几个待审核价格
	content += "\n**最近待审核价格（最多显示5个）：**\n"
	maxDisplay := 5
	if len(pendingPrices) < maxDisplay {
		maxDisplay = len(pendingPrices)
	}

	for i := 0; i < maxDisplay; i++ {
		price := pendingPrices[i]
		var provider models.Provider
		channelType := getChannelType(price)
		if err := database.DB.Where("id = ?", channelType).First(&provider).Error; err != nil {
			provider.Name = fmt.Sprintf("厂商ID:%d", channelType)
		}

		content += fmt.Sprintf("%d. **%s** (%s) - 创建者：%s\n",
			i+1,
			getDisplayModel(price),
			provider.Name,
			price.CreatedBy)
	}

	if len(pendingPrices) > maxDisplay {
		content += fmt.Sprintf("\n...还有 %d 个价格等待审核", len(pendingPrices)-maxDisplay)
	}

	content += "\n\n⏰ 请及时处理待审核价格！"

	// 发送卡片通知
	return webhook.SendPendingPricesDetailedNotification(content, len(pendingPrices))
}

// getChannelType 获取厂商类型
func getChannelType(price models.Price) uint {
	if price.TempChannelType != nil {
		return *price.TempChannelType
	}
	return price.ChannelType
}

// getDisplayModel 获取显示用的模型名称
func getDisplayModel(price models.Price) string {
	if price.TempModel != nil {
		return *price.TempModel
	}
	return price.Model
}
