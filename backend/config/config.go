package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath         string
	ServerPort     string
	AdminUsernames []string
}

func LoadConfig() (*Config, error) {
	// 确保数据目录存在
	dbDir := "./data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// 尝试从 data 目录加载 .env 文件
	envPath := filepath.Join(dbDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("Warning: .env file not found in data directory: %v\n", err)
		// 如果 data/.env 不存在，尝试加载项目根目录的 .env
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: .env file not found in root directory: %v\n", err)
		}
	}

	// 从环境变量获取管理员用户名列表，多个用户名用逗号分隔
	rawAdminUsernames := os.Getenv("ADMIN_USERNAMES")
	fmt.Printf("Raw ADMIN_USERNAMES from env: %q\n", rawAdminUsernames)

	adminUsernames := strings.Split(getEnv("ADMIN_USERNAMES", "admin"), ",")
	// 去除每个用户名的空白字符
	for i, username := range adminUsernames {
		adminUsernames[i] = strings.TrimSpace(username)
	}

	fmt.Printf("Processed admin usernames: %v\n", adminUsernames)

	config := &Config{
		DBPath:         filepath.Join(dbDir, "aimodels.db"),
		ServerPort:     getEnv("PORT", "8080"),
		AdminUsernames: adminUsernames,
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
