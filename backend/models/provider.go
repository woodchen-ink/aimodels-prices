package models

import "time"

type Provider struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
}

// CreateProviderTableSQL 返回创建供应商表的 SQL
func CreateProviderTableSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS provider (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		icon TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT NOT NULL
	)`
}
