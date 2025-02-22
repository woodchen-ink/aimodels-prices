package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"aimodels-prices/config"
	"aimodels-prices/models"
)

// DB 是数据库连接的全局实例
var DB *sql.DB

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
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	// 创建表结构
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	return nil
}

// createTables 创建数据库表
func createTables() error {
	// 创建用户表
	if _, err := DB.Exec(models.CreateUserTableSQL()); err != nil {
		log.Printf("Failed to create user table: %v", err)
		return err
	}

	// 创建会话表
	if _, err := DB.Exec(models.CreateSessionTableSQL()); err != nil {
		log.Printf("Failed to create session table: %v", err)
		return err
	}

	// 创建模型厂商表
	if _, err := DB.Exec(models.CreateProviderTableSQL()); err != nil {
		log.Printf("Failed to create provider table: %v", err)
		return err
	}

	// 创建模型类型表
	if _, err := DB.Exec(models.CreateModelTypeTableSQL()); err != nil {
		log.Printf("Failed to create model_type table: %v", err)
		return err
	}

	// 创建价格表
	if _, err := DB.Exec(models.CreatePriceTableSQL()); err != nil {
		log.Printf("Failed to create price table: %v", err)
		return err
	}

	return nil
}
