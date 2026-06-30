package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionMailBox  = "mail_boxes"
	CollectionMailBody = "mail_bodies"
)

// MailBox 玩家邮箱索引（MongoDB）
type MailBox struct {
	ID        primitive.ObjectID `bson:"_id"`
	PlayerID  string             `bson:"player_id"`
	MailID    string             `bson:"mail_id"`
	Status    int32              `bson:"status"` // 0=未读, 1=已读, 2=已领取, 3=已删除
	ReadAt    *time.Time         `bson:"read_at,omitempty"`
	ClaimedAt *time.Time         `bson:"claimed_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	ExpireAt  *time.Time         `bson:"expire_at,omitempty"`
}

// MailBody 邮件正文（MongoDB）
type MailBody struct {
	ID          primitive.ObjectID `bson:"_id"`
	MailID      string             `bson:"mail_id"`
	SenderID    string             `bson:"sender_id"`
	SenderName  string             `bson:"sender_name"`
	Title       string             `bson:"title"`
	Content     string             `bson:"content"`
	MailType    int32              `bson:"mail_type"` // 0=系统, 1=玩家
	Attachments []Attachment       `bson:"attachments,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
}

// Attachment 附件
type Attachment struct {
	Type      string `bson:"type"`       // currency/asset
	ID        string `bson:"id"`         // 货币类型ID或资产ID
	Amount    int32  `bson:"amount"`     // 数量
	AssetType int32  `bson:"asset_type"` // 资产类型（asset时有效）
}
