package model

import (
	"fight-game/pkg/common/model"
)

type User struct {
	model.Base
	Username string `gorm:"size:64"`          // 用户名
	Password string `gorm:"not null;size:64"` // 密码
	Email    string `gorm:"size:128"`         // 邮箱
	Phone    string `gorm:"size:16"`          // 手机号
}

func (User) TableName() string {
	return "user"
}
