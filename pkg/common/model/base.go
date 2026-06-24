package model

import "time"

type Base struct {
	Id        string    `gorm:"primarykey"` // 用户ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}
