package init

import (
	"log"
)

// RunInitTasks 运行所有初始化任务
func RunInitTasks() {
	// 检查并处理重复的模型名称
	if err := CheckDuplicateModelNames(); err != nil {
		log.Printf("检查重复模型名称时发生错误: %v", err)
	}

	// 在此处添加其他初始化任务
	// ...
}
