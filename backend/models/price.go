package models

import (
	"time"

	"gorm.io/gorm"
)

type Price struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Model             string         `json:"model" gorm:"not null;index:idx_model_channel"`
	ModelType         string         `json:"model_type" gorm:"not null;index:idx_model_type"` // text2text, text2image, etc.
	BillingType       string         `json:"billing_type" gorm:"not null"`                    // tokens or times
	ChannelType       uint           `json:"channel_type" gorm:"not null;index:idx_model_channel"`
	Currency          string         `json:"currency" gorm:"not null"` // USD or CNY
	InputPrice        float64        `json:"input_price" gorm:"not null"`
	OutputPrice       float64        `json:"output_price" gorm:"not null"`
	InputAudioTokens  *float64       `json:"input_audio_tokens,omitempty"`  // 音频输入价格
	OutputAudioTokens *float64       `json:"output_audio_tokens,omitempty"` // 音频输出价格
	CachedTokens      *float64       `json:"cached_tokens,omitempty"`       // 缓存价格
	CachedReadTokens  *float64       `json:"cached_read_tokens,omitempty"`  // 缓存读取价格
	CachedWriteTokens *float64       `json:"cached_write_tokens,omitempty"` // 缓存写入价格
	ReasoningTokens   *float64       `json:"reasoning_tokens,omitempty"`    // 推理价格
	InputTextTokens   *float64       `json:"input_text_tokens,omitempty"`   // 输入文本价格
	OutputTextTokens  *float64       `json:"output_text_tokens,omitempty"`  // 输出文本价格
	InputImageTokens  *float64       `json:"input_image_tokens,omitempty"`  // 输入图片价格
	OutputImageTokens *float64       `json:"output_image_tokens,omitempty"` // 输出图片价格
	PriceSource       string         `json:"price_source" gorm:"not null"`
	Status            string         `json:"status" gorm:"not null;default:pending;index:idx_status"` // pending, approved, rejected
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime;index:idx_created_at"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy         string         `json:"created_by" gorm:"not null"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	// 临时字段，用于存储待审核的更新
	TempModel             *string  `json:"temp_model,omitempty" gorm:"column:temp_model"`
	TempModelType         *string  `json:"temp_model_type,omitempty" gorm:"column:temp_model_type"`
	TempBillingType       *string  `json:"temp_billing_type,omitempty" gorm:"column:temp_billing_type"`
	TempChannelType       *uint    `json:"temp_channel_type,omitempty" gorm:"column:temp_channel_type"`
	TempCurrency          *string  `json:"temp_currency,omitempty" gorm:"column:temp_currency"`
	TempInputPrice        *float64 `json:"temp_input_price,omitempty" gorm:"column:temp_input_price"`
	TempOutputPrice       *float64 `json:"temp_output_price,omitempty" gorm:"column:temp_output_price"`
	TempInputAudioTokens  *float64 `json:"temp_input_audio_tokens,omitempty"`
	TempOutputAudioTokens *float64 `json:"temp_output_audio_tokens,omitempty"`
	TempCachedTokens      *float64 `json:"temp_cached_tokens,omitempty"`
	TempCachedReadTokens  *float64 `json:"temp_cached_read_tokens,omitempty"`
	TempCachedWriteTokens *float64 `json:"temp_cached_write_tokens,omitempty"`
	TempReasoningTokens   *float64 `json:"temp_reasoning_tokens,omitempty"`
	TempInputTextTokens   *float64 `json:"temp_input_text_tokens,omitempty"`
	TempOutputTextTokens  *float64 `json:"temp_output_text_tokens,omitempty"`
	TempInputImageTokens  *float64 `json:"temp_input_image_tokens,omitempty"`
	TempOutputImageTokens *float64 `json:"temp_output_image_tokens,omitempty"`
	TempPriceSource       *string  `json:"temp_price_source,omitempty" gorm:"column:temp_price_source"`
	UpdatedBy             *string  `json:"updated_by,omitempty" gorm:"column:updated_by"`
}

// TableName 指定表名
func (Price) TableName() string {
	return "price"
}
