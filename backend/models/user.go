package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"not null;unique"`
	Email     string         `json:"email" gorm:"not null;type:varchar(191)"`
	Role      string         `json:"role" gorm:"not null;default:user"`    // admin or user (legacy)
	Groups    string         `json:"groups" gorm:"type:text;default:'t0'"` // CZL Connect权限组: t0,t1,t2,t3,t4,t5,viewer,admin
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Session struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定User表名
func (User) TableName() string {
	return "user"
}

// TableName 指定Session表名
func (Session) TableName() string {
	return "session"
}
