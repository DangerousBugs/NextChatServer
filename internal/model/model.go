package model

import (
	"time"

	"gorm.io/gorm"
)

// User 账号表
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"size:32;uniqueIndex;not null"`
	Email     string         `gorm:"size:128;uniqueIndex"`
	Password  string         `gorm:"size:64;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`      // 软删
	Version   uint           `gorm:"default:1"`  // 乐观锁
	Profile   Profile        `gorm:"constraint:OnDelete:CASCADE"` // 一对一
	Messages  []Message
}

// Profile 资料表（一对一）
type Profile struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"uniqueIndex"` // 唯一外键
	Bio    string `gorm:"type:text"`
	Avatar string `gorm:"size:256"`
}

// Message 消息表（多对一）
type Message struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Content   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"index"`
}
