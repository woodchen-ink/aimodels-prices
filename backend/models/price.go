package models

import (
	"time"
)

type Price struct {
	ID          uint      `json:"id"`
	Model       string    `json:"model"`
	ModelType   string    `json:"model_type"`   // text2text, text2image, etc.
	BillingType string    `json:"billing_type"` // tokens or times
	ChannelType string    `json:"channel_type"`
	Currency    string    `json:"currency"` // USD or CNY
	InputPrice  float64   `json:"input_price"`
	OutputPrice float64   `json:"output_price"`
	PriceSource string    `json:"price_source"`
	Status      string    `json:"status"` // pending, approved, rejected
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	// 临时字段，用于存储待审核的更新
	TempModel       *string  `json:"temp_model,omitempty"`
	TempModelType   *string  `json:"temp_model_type,omitempty"`
	TempBillingType *string  `json:"temp_billing_type,omitempty"`
	TempChannelType *string  `json:"temp_channel_type,omitempty"`
	TempCurrency    *string  `json:"temp_currency,omitempty"`
	TempInputPrice  *float64 `json:"temp_input_price,omitempty"`
	TempOutputPrice *float64 `json:"temp_output_price,omitempty"`
	TempPriceSource *string  `json:"temp_price_source,omitempty"`
	UpdatedBy       *string  `json:"updated_by,omitempty"`
}

// CreatePriceTableSQL 返回创建价格表的 SQL
func CreatePriceTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS price (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		model VARCHAR(255) NOT NULL,
		model_type VARCHAR(50) NOT NULL,
		billing_type VARCHAR(50) NOT NULL,
		channel_type VARCHAR(50) NOT NULL,
		currency VARCHAR(10) NOT NULL,
		input_price DECIMAL(10,6) NOT NULL,
		output_price DECIMAL(10,6) NOT NULL,
		price_source VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		created_by VARCHAR(255) NOT NULL,
		temp_model VARCHAR(255),
		temp_model_type VARCHAR(50),
		temp_billing_type VARCHAR(50),
		temp_channel_type VARCHAR(50),
		temp_currency VARCHAR(10),
		temp_input_price DECIMAL(10,6),
		temp_output_price DECIMAL(10,6),
		temp_price_source VARCHAR(255),
		updated_by VARCHAR(255),
		FOREIGN KEY (channel_type) REFERENCES provider(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
}
