package model

import "fight-game/pkg/common/model"

// Player 玩家基础信息表
type Player struct {
	model.Base
	UserID    string `gorm:"primaryKey;column:player_id;type:varchar(64)"`
	Nickname  string `gorm:"column:nickname;type:varchar(32);not null;default:''"`
	Level     int32  `gorm:"column:level;not null;default:1"`
	Exp       int64  `gorm:"column:exp;not null;default:0"`
	AvatarURL string `gorm:"column:avatar_url;type:varchar(256);not null;default:''"`
	Signature string `gorm:"column:signature;type:varchar(128);not null;default:''"`
}

func (Player) TableName() string {
	return "player"
}
