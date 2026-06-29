package model

import "fight-game/pkg/common/model"

// PlayerCurrency 玩家货币表
type PlayerCurrency struct {
	model.Base
	PlayerId     string `gorm:"column:player_id;type:varchar(64);not null;index:idx_player_currency"`
	CurrencyType int16  `gorm:"column:currency_type;not null"`
	Amount       int64  `gorm:"column:amount;not null;default:0"`
	Version      int32  `gorm:"column:version;not null;default:0"` // 乐观锁
}

func (PlayerCurrency) TableName() string {
	return "player_currencies"
}
