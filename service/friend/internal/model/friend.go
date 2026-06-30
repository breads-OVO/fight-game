package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Friend 好友关系表
type Friend struct {
	ID        uint      `gorm:"primarykey;autoIncrement"`
	PlayerID  string    `gorm:"column:player_id;type:varchar(64);not null;index:idx_player;uniqueIndex:uk_friend"`
	FriendID  string    `gorm:"column:friend_id;type:varchar(64);not null;uniqueIndex:uk_friend"`
	Status    int32     `gorm:"column:status;type:tinyint;not null;default:0;comment:'0=待确认,1=已接受'"` // 0=pending, 1=accepted
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Friend) TableName() string {
	return "friends"
}

// ChatMessage 聊天消息（MongoDB）
type ChatMessage struct {
	ID           primitive.ObjectID `bson:"_id"`
	MessageID    string             `bson:"message_id"`
	SenderID     string             `bson:"sender_id"`
	SenderName   string             `bson:"sender_name"`
	ReceiverID   string             `bson:"receiver_id"`
	ChatType     int32              `bson:"chat_type"` // 0=私聊, 1=房间, 2=公会
	Content      string             `bson:"content"`
	Participants []string           `bson:"participants"` // 参与者ID列表（用于查询）
	CreatedAt    time.Time          `bson:"created_at"`
}

// Collection names
const (
	CollectionChatMessages = "chat_messages"
)
