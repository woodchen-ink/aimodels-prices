package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// ModelType 模型类型结构
type ModelType struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

func main() {
	// 确保数据目录存在
	dbDir := "./data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 连接数据库
	dbPath := filepath.Join(dbDir, "aimodels.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 创建model_type表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS model_type (
			key TEXT PRIMARY KEY,
			label TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("创建model_type表失败: %v", err)
	}

	// 初始化默认的模型类型
	defaultTypes := []ModelType{
		{Key: "text2text", Label: "文生文"},
		{Key: "text2image", Label: "文生图"},
		{Key: "text2speech", Label: "文生音"},
		{Key: "speech2text", Label: "音生文"},
		{Key: "image2text", Label: "图生文"},
		{Key: "embedding", Label: "向量"},
		{Key: "other", Label: "其他"},
	}

	// 插入默认类型
	for _, t := range defaultTypes {
		_, err = db.Exec(`
			INSERT OR IGNORE INTO model_type (key, label)
			VALUES (?, ?)
		`, t.Key, t.Label)
		if err != nil {
			log.Printf("插入默认类型失败 %s: %v", t.Key, err)
		}
	}

	// 检查model_type列是否存在
	var hasModelType bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM pragma_table_info('price') 
		WHERE name = 'model_type'
	`).Scan(&hasModelType)
	if err != nil {
		log.Fatalf("检查model_type列失败: %v", err)
	}

	// 如果model_type列不存在,则添加它
	if !hasModelType {
		log.Println("开始添加model_type列...")

		// 开始事务
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("开始事务失败: %v", err)
		}

		// 添加model_type列
		_, err = tx.Exec(`ALTER TABLE price ADD COLUMN model_type TEXT`)
		if err != nil {
			tx.Rollback()
			log.Fatalf("添加model_type列失败: %v", err)
		}

		// 添加temp_model_type列
		_, err = tx.Exec(`ALTER TABLE price ADD COLUMN temp_model_type TEXT`)
		if err != nil {
			tx.Rollback()
			log.Fatalf("添加temp_model_type列失败: %v", err)
		}

		// 根据模型名称推断类型并更新
		rows, err := tx.Query(`SELECT id, model FROM price`)
		if err != nil {
			tx.Rollback()
			log.Fatalf("查询价格数据失败: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var model string
			if err := rows.Scan(&id, &model); err != nil {
				tx.Rollback()
				log.Fatalf("读取行数据失败: %v", err)
			}

			// 根据模型名称推断类型
			modelType := inferModelType(model)

			// 更新model_type
			_, err = tx.Exec(`UPDATE price SET model_type = ? WHERE id = ?`, modelType, id)
			if err != nil {
				tx.Rollback()
				log.Fatalf("更新model_type失败: %v", err)
			}
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			log.Fatalf("提交事务失败: %v", err)
		}

		log.Println("成功添加并更新model_type列")
	} else {
		log.Println("model_type列已存在,无需迁移")
	}
}

// inferModelType 根据模型名称推断模型类型
func inferModelType(model string) string {
	model = strings.ToLower(model)

	switch {
	case strings.Contains(model, "gpt") ||
		strings.Contains(model, "llama") ||
		strings.Contains(model, "claude") ||
		strings.Contains(model, "palm") ||
		strings.Contains(model, "gemini") ||
		strings.Contains(model, "qwen") ||
		strings.Contains(model, "chatglm"):
		return "text2text"

	case strings.Contains(model, "dall-e") ||
		strings.Contains(model, "stable") ||
		strings.Contains(model, "midjourney") ||
		strings.Contains(model, "sd") ||
		strings.Contains(model, "diffusion"):
		return "text2image"

	case strings.Contains(model, "whisper") ||
		strings.Contains(model, "speech") ||
		strings.Contains(model, "tts"):
		return "text2speech"

	case strings.Contains(model, "embedding") ||
		strings.Contains(model, "ada") ||
		strings.Contains(model, "text-embedding"):
		return "embedding"

	default:
		return "other"
	}
}
