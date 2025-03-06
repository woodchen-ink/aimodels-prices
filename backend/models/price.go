package models

import (
	"time"

	"gorm.io/gorm"
)

type Price struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Model       string         `json:"model" gorm:"not null"`
	ModelType   string         `json:"model_type" gorm:"not null"`   // text2text, text2image, etc.
	BillingType string         `json:"billing_type" gorm:"not null"` // tokens or times
	ChannelType uint           `json:"channel_type" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"not null"` // USD or CNY
	InputPrice  float64        `json:"input_price" gorm:"not null"`
	OutputPrice float64        `json:"output_price" gorm:"not null"`
	PriceSource string         `json:"price_source" gorm:"not null"`
	Status      string         `json:"status" gorm:"not null;default:pending"` // pending, approved, rejected
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy   string         `json:"created_by" gorm:"not null"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	// 临时字段，用于存储待审核的更新
	TempModel       *string  `json:"temp_model,omitempty" gorm:"column:temp_model"`
	TempModelType   *string  `json:"temp_model_type,omitempty" gorm:"column:temp_model_type"`
	TempBillingType *string  `json:"temp_billing_type,omitempty" gorm:"column:temp_billing_type"`
	TempChannelType *uint    `json:"temp_channel_type,omitempty" gorm:"column:temp_channel_type"`
	TempCurrency    *string  `json:"temp_currency,omitempty" gorm:"column:temp_currency"`
	TempInputPrice  *float64 `json:"temp_input_price,omitempty" gorm:"column:temp_input_price"`
	TempOutputPrice *float64 `json:"temp_output_price,omitempty" gorm:"column:temp_output_price"`
	TempPriceSource *string  `json:"temp_price_source,omitempty" gorm:"column:temp_price_source"`
	UpdatedBy       *string  `json:"updated_by,omitempty" gorm:"column:updated_by"`
}

// TableName 指定表名
func (Price) TableName() string {
	return "price"
}
