package models

import (
	"time"

	"gorm.io/gorm"
)

// ModelType 模型类型结构
type ModelType struct {
	TypeKey   string         `json:"key" gorm:"primaryKey;column:type_key"`
	TypeLabel string         `json:"label" gorm:"column:type_label;not null"`
	SortOrder int            `json:"sort_order" gorm:"column:sort_order;default:0"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ModelType) TableName() string {
	return "model_type"
}
