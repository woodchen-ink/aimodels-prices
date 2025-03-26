package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"

	openrouter_api "aimodels-prices/cron/openrouter-api"
	siliconflow_api "aimodels-prices/cron/siliconflow-api"
)

var cronScheduler *cron.Cron

// Init 初始化并启动所有定时任务
func Init() {
	log.Println("初始化定时任务...")

	// 创建一个新的cron调度器，使用秒级精度
	cronScheduler = cron.New(cron.WithSeconds())

	// 注册价格获取任务
	// 每4小时执行一次
	_, err := cronScheduler.AddFunc("0 */4 * * *", func() {
		if err := openrouter_api.FetchAndSavePrices(); err != nil {
			log.Printf("OpenRouter价格获取任务执行失败: %v", err)
		}

		time.Sleep(3 * time.Second)

		if err := openrouter_api.UpdateOtherPrices(); err != nil {
			log.Printf("其他厂商价格更新任务执行失败: %v", err)
		}

		time.Sleep(3 * time.Second)

		if err := siliconflow_api.UpdateSiliconFlowPrices(); err != nil {
			log.Printf("SiliconFlow价格更新任务执行失败: %v", err)
		}
	})

	if err != nil {
		log.Printf("注册价格更新定时任务失败: %v", err)
	}

	// 启动定时任务
	cronScheduler.Start()
	log.Println("定时任务已启动")

	// 立即执行一次价格获取任务
	go func() {
		// 等待几秒钟，确保应用程序和数据库已完全初始化
		time.Sleep(5 * time.Second)
		log.Println("立即执行OpenRouter价格获取任务...")
		if err := openrouter_api.FetchAndSavePrices(); err != nil {
			log.Printf("初始OpenRouter价格获取任务执行失败: %v", err)
		}

		// 等待几秒后执行其他厂商价格更新任务
		time.Sleep(3 * time.Second)
		log.Println("立即执行其他厂商价格更新任务...")
		if err := openrouter_api.UpdateOtherPrices(); err != nil {
			log.Printf("初始其他厂商价格更新任务执行失败: %v", err)
		}

		// 等待几秒后执行SiliconFlow价格更新任务
		time.Sleep(3 * time.Second)
		log.Println("立即执行SiliconFlow价格更新任务...")
		if err := siliconflow_api.UpdateSiliconFlowPrices(); err != nil {
			log.Printf("初始SiliconFlow价格更新任务执行失败: %v", err)
		}
	}()
}

// StopCronJobs 停止所有定时任务
func StopCronJobs() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("定时任务已停止")
	}
}
