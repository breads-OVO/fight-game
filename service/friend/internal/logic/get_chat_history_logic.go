package logic

import (
	"context"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetChatHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatHistoryLogic {
	return &GetChatHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetChatHistory 获取聊天记录
func (l *GetChatHistoryLogic) GetChatHistory(in *friend.GetChatHistoryRequest) (*friend.GetChatHistoryResponse, error) {
	page := int64(in.Page)
	pageSize := int64(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	coll := l.svcCtx.MongoDB.Collection(model.CollectionChatMessages)

	var filter bson.M

	switch in.ChatType {
	case friend.ChatType_PRIVATE:
		// 私聊：查询双方参与的会话
		filter = bson.M{
			"chat_type": int32(friend.ChatType_PRIVATE),
			"participants": bson.M{
				"$all": []string{in.PlayerId, in.TargetId},
			},
		}
	case friend.ChatType_ROOM:
		filter = bson.M{
			"chat_type":   int32(friend.ChatType_ROOM),
			"receiver_id": in.TargetId,
		}
	case friend.ChatType_GUILD:
		filter = bson.M{
			"chat_type":   int32(friend.ChatType_GUILD),
			"receiver_id": in.TargetId,
		}
	default:
		filter = bson.M{
			"chat_type": int32(friend.ChatType_PRIVATE),
			"participants": bson.M{
				"$all": []string{in.PlayerId, in.TargetId},
			},
		}
	}

	// 查询总数
	total, err := coll.CountDocuments(l.ctx, filter)
	if err != nil {
		logx.Errorf("CountDocuments error: %v", err)
		return nil, err
	}

	// 分页查询（按时间倒序）
	skip := (page - 1) * pageSize
	cursor, err := coll.Find(l.ctx, filter,
		options.Find().SetSkip(skip).SetLimit(pageSize).
			SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		logx.Errorf("Find chat history error: %v", err)
		return nil, err
	}
	defer cursor.Close(l.ctx)

	var messages []model.ChatMessage
	if err := cursor.All(l.ctx, &messages); err != nil {
		logx.Errorf("Cursor.All error: %v", err)
		return nil, err
	}

	// 反转时间顺序（按时间正序返回）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// 组装响应
	chatMessages := make([]*friend.ChatMessage, 0, len(messages))
	for _, m := range messages {
		chatMessages = append(chatMessages, &friend.ChatMessage{
			MessageId:  m.MessageID,
			SenderId:   m.SenderID,
			SenderName: m.SenderName,
			ReceiverId: m.ReceiverID,
			ChatType:   friend.ChatType(m.ChatType),
			Content:    m.Content,
			CreatedAt:  m.CreatedAt.Unix(),
		})
	}

	return &friend.GetChatHistoryResponse{
		Messages: chatMessages,
		Total:    int32(total),
	}, nil
}
