package logic

import (
	"context"
	"errors"
	"time"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AddFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendLogic {
	return &AddFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddFriend 添加好友
func (l *AddFriendLogic) AddFriend(in *friend.AddFriendRequest) (*friend.AddFriendResponse, error) {
	if in.PlayerId == in.FriendId {
		return &friend.AddFriendResponse{
			Success: false,
			Message: "不能添加自己为好友",
		}, nil
	}

	now := time.Now()

	// 检查是否已存在好友关系
	var existing model.Friend
	err := l.svcCtx.DB.Where(
		"(player_id = ? AND friend_id = ?) OR (player_id = ? AND friend_id = ?)",
		in.PlayerId, in.FriendId, in.FriendId, in.PlayerId,
	).First(&existing).Error

	if err == nil {
		// 已存在记录
		if existing.Status == 1 {
			return &friend.AddFriendResponse{
				Success: false,
				Message: "已经是好友了",
			}, nil
		}
		if existing.Status == 0 {
			return &friend.AddFriendResponse{
				Success: false,
				Message: "好友请求已发送，请等待对方确认",
			}, nil
		}
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logx.Errorf("AddFriend query error: %v", err)
		return nil, err
	}

	// 创建好友请求记录
	friendRecord := model.Friend{
		PlayerID: in.PlayerId,
		FriendID: in.FriendId,
		Status:   0, // 待确认
	}
	friendRecord.CreatedAt = now

	if err := l.svcCtx.DB.Create(&friendRecord).Error; err != nil {
		logx.Errorf("AddFriend create error: %v", err)
		return nil, err
	}

	logx.Infof("Friend request sent: from=%s, to=%s", in.PlayerId, in.FriendId)
	return &friend.AddFriendResponse{
		Success: true,
		Message: "好友请求已发送",
	}, nil
}
