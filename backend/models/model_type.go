package models

// ModelType 模型类型结构
type ModelType struct {
	TypeKey   string `json:"key"`
	TypeLabel string `json:"label"`
}

// CreateModelTypeTableSQL 返回创建模型类型表的 SQL
func CreateModelTypeTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS model_type (
		type_key VARCHAR(50) PRIMARY KEY,
		type_label VARCHAR(255) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
}
