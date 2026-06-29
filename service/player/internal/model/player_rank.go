package model

import "fight-game/pkg/common/model"

// PlayerRank 玩家段位表
type PlayerRank struct {
	model.Base
	PlayerId  string `gorm:"column:player_id;type:varchar(64);not null;index:idx_player_currency"`
	Rating    int32  `gorm:"column:rating;not null"`
	TotalGame int32  `gorm:"column:total_game;not null;default:0"`
	Win       int32  `gorm:"column:win;not null;default:0"`
	Lose      int32  `gorm:"column:lose;not null;default:0"`
	Season    string `gorm:"column:season;type:varchar(32);not null;default:''"`
}

func (PlayerRank) TableName() string {
	return "player_rank"
}
