package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath     string
	ServerPort string
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

	config := &Config{
		DBPath:     filepath.Join(dbDir, "aimodels.db"),
		ServerPort: getEnv("PORT", "8080"),
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
