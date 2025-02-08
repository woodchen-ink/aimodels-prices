package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"aimodels-prices/models"
)

// DB 是数据库连接的全局实例
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return err
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	// 创建表结构
	if err = createTables(); err != nil {
		return err
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

	// 创建供应商表
	if _, err := DB.Exec(models.CreateProviderTableSQL()); err != nil {
		log.Printf("Failed to create provider table: %v", err)
		return err
	}

	// 创建价格表
	if _, err := DB.Exec(models.CreatePriceTableSQL()); err != nil {
		log.Printf("Failed to create price table: %v", err)
		return err
	}

	return nil
}
