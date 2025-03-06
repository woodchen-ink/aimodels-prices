package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"aimodels-prices/config"
	"aimodels-prices/models"
)

// DB 是数据库连接的全局实例
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	var err error

	// 构建MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// 连接MySQL
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// 获取底层的SQL DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	// 自动迁移表结构
	if err = migrateModels(); err != nil {
		return fmt.Errorf("failed to migrate models: %v", err)
	}

	return nil
}

// migrateModels 自动迁移模型到数据库表
func migrateModels() error {
	// 自动迁移模型
	if err := DB.AutoMigrate(
		&models.ModelType{},
		&models.Price{},
		&models.Provider{},
		&models.User{},
		&models.Session{},
	); err != nil {
		log.Printf("Failed to migrate tables: %v", err)
		return err
	}

	// 这里可以添加其他模型的迁移
	// 例如：DB.AutoMigrate(&models.User{})

	return nil
}
