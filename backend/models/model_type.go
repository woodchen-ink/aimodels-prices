package models

// ModelType 模型类型结构
type ModelType struct {
	TypeKey   string `json:"key"`
	TypeLabel string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

// CreateModelTypeTableSQL 返回创建模型类型表的 SQL
func CreateModelTypeTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS model_type (
		type_key VARCHAR(50) PRIMARY KEY,
		type_label VARCHAR(255) NOT NULL,
		sort_order INT NOT NULL DEFAULT 0
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
}
