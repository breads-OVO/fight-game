package logic

import (
	"context"
	"fmt"
	"time"

	"fight-game/pb/common"
	"fight-game/pb/friend"
	"fight-game/pb/gateway"
	"fight-game/pb/player"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
)

type SendChatMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendChatMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendChatMessageLogic {
	return &SendChatMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendChatMessageLogic) SendChatMessage(in *friend.SendChatMessageRequest) (*friend.SendChatMessageResponse, error) {
	now := time.Now()

	messageID := uuid.New().String()

	// 构建参与者列表（用于查询聊天历史时能通过任一方查到）
	participants := []string{}
	switch in.ChatType {
	case friend.ChatType_PRIVATE:
		participants = []string{in.SenderId, in.ReceiverId}
	case friend.ChatType_ROOM:
		// 房间消息，参与者由房间决定，暂不处理
	default:
		participants = []string{in.SenderId}
	}

	// 从 Player 服务获取发送者昵称
	senderName := l.fetchSenderName(in.SenderId)

	// 组装消息
	msg := model.ChatMessage{
		ID:           primitive.NewObjectID(),
		MessageID:    messageID,
		SenderID:     in.SenderId,
		SenderName:   senderName,
		ReceiverID:   in.ReceiverId,
		ChatType:     int32(in.ChatType),
		Content:      in.Content,
		Participants: participants,
		CreatedAt:    now,
	}

	// 写入 MongoDB
	coll := l.svcCtx.MongoDB.Collection(model.CollectionChatMessages)
	if _, err := coll.InsertOne(l.ctx, msg); err != nil {
		logx.Errorf("SendChatMessage insert error: %v", err)
		return nil, err
	}

	// 如果是私聊且接收者在线，通过 Gateway PushService 推送消息给接收者
	if in.ChatType == friend.ChatType_PRIVATE && in.ReceiverId != "" {
		go l.pushToReceiver(in.SenderId, in.ReceiverId, senderName, in.Content, now)
	}

	logx.Infof("Chat message sent: sender=%s, receiver=%s, type=%d",
		in.SenderId, in.ReceiverId, in.ChatType)

	return &friend.SendChatMessageResponse{
		Message: &friend.ChatMessage{
			MessageId:  messageID,
			SenderId:   in.SenderId,
			SenderName: senderName,
			ReceiverId: in.ReceiverId,
			ChatType:   in.ChatType,
			Content:    in.Content,
			CreatedAt:  now.Unix(),
		},
	}, nil
}

// fetchSenderName 从 Player 服务获取发送者昵称
func (l *SendChatMessageLogic) fetchSenderName(senderId string) string {
	resp, err := l.svcCtx.PlayerClient.GetProfile(l.ctx, &player.GetProfileRequest{
		PlayerId: senderId,
	})
	if err != nil {
		logx.Errorf("fetchSenderName GetProfile error: player=%s, err=%v", senderId, err)
		return senderId
	}
	if resp == nil || resp.Profile == nil {
		return senderId
	}
	return resp.Profile.Nickname
}

// pushToReceiver 通过 Gateway PushService 推送聊天消息给在线接收者
func (l *SendChatMessageLogic) pushToReceiver(senderId, receiverId, senderName, content string, createdAt time.Time) {
	// 先检查 Redis 在线状态
	key := fmt.Sprintf("social:online:%s", receiverId)
	exists, err := l.svcCtx.RedisClient.ExistsCtx(l.ctx, key)
	if err != nil || !exists {
		return // 离线，不推送
	}

	// 构建推送消息体
	chatMsg := &friend.ChatMessage{
		MessageId:  uuid.New().String(),
		SenderId:   senderId,
		SenderName: senderName,
		ReceiverId: receiverId,
		ChatType:   friend.ChatType_PRIVATE,
		Content:    content,
		CreatedAt:  createdAt.Unix(),
	}

	body, err := proto.Marshal(chatMsg)
	if err != nil {
		logx.Errorf("pushToReceiver marshal error: %v", err)
		return
	}

	// 调用 Gateway PushService 推送（使用 PUSH_CHAT_MESSAGE 消息类型）
	_, err = l.svcCtx.PushClient.PushMessage(l.ctx, &gateway.PushMessageRequest{
		PlayerId: receiverId,
		MsgType:  int32(common.WSMsgType_MSG_PUSH_CHAT_MESSAGE),
		Body:     body,
	})
	if err != nil {
		logx.Errorf("pushToReceiver push failed: receiver=%s, err=%v", receiverId, err)
		return
	}

	logx.Infof("Chat message pushed to receiver=%s, sender=%s", receiverId, senderId)
}
