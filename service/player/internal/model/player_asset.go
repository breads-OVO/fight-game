package model

import "fight-game/pkg/common/model"

// PlayerAsset 玩家资产表
type PlayerAsset struct {
	model.Base
	PlayerId  string `gorm:"column:player_id;type:varchar(64);not null;index:idx_player"`
	AssetId   string `gorm:"column:asset_id;type:varchar(64);not null"`
	AssetType int16  `gorm:"column:asset_type;not null"`
	Quantity  int32  `gorm:"column:quantity;not null;default:1"`
	Status    int16  `gorm:"column:status;not null;default:0"`    // 0=NORMAL 1=LOCKED 2=TEMP
	ExpireAt  int64  `gorm:"column:expire_at;not null;default:0"` // 0=永久
}

func (PlayerAsset) TableName() string {
	return "player_asset"
}
