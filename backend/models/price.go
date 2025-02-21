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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model TEXT NOT NULL,
		model_type TEXT NOT NULL,
		billing_type TEXT NOT NULL,
		channel_type TEXT NOT NULL,
		currency TEXT NOT NULL,
		input_price REAL NOT NULL,
		output_price REAL NOT NULL,
		price_source TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT NOT NULL,
		temp_model TEXT,
		temp_model_type TEXT,
		temp_billing_type TEXT,
		temp_channel_type TEXT,
		temp_currency TEXT,
		temp_input_price REAL,
		temp_output_price REAL,
		temp_price_source TEXT,
		updated_by TEXT,
		FOREIGN KEY (channel_type) REFERENCES provider(id)
	)`
}
