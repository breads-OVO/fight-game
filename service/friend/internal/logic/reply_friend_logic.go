package logic

import (
	"context"
	"time"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplyFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplyFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyFriendLogic {
	return &ReplyFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReplyFriend 回复好友请求
func (l *ReplyFriendLogic) ReplyFriend(in *friend.ReplyFriendRequest) (*friend.ReplyFriendResponse, error) {
	now := time.Now()

	if in.Accept {
		// 接受好友请求：找到待确认的记录并更新状态
		// 同时添加双向关系（对方添加我的那条记录）
		result := l.svcCtx.DB.Model(&model.Friend{}).
			Where("player_id = ? AND friend_id = ? AND status = ?", in.FriendId, in.PlayerId, 0).
			Updates(map[string]interface{}{
				"status":     1, // 已接受
				"updated_at": now,
			})

		if result.Error != nil {
			logx.Errorf("ReplyFriend accept error: %v", result.Error)
			return &friend.ReplyFriendResponse{Success: false}, result.Error
		}

		if result.RowsAffected == 0 {
			return &friend.ReplyFriendResponse{
				Success: false,
			}, nil
		}

		logx.Infof("Friend request accepted: from=%s, to=%s", in.FriendId, in.PlayerId)
	} else {
		// 拒绝好友请求：删除记录
		result := l.svcCtx.DB.Where(
			"player_id = ? AND friend_id = ? AND status = ?",
			in.FriendId, in.PlayerId, 0,
		).Delete(&model.Friend{})

		if result.Error != nil {
			logx.Errorf("ReplyFriend reject error: %v", result.Error)
			return &friend.ReplyFriendResponse{Success: false}, result.Error
		}

		logx.Infof("Friend request rejected: from=%s, to=%s", in.FriendId, in.PlayerId)
	}

	return &friend.ReplyFriendResponse{
		Success: true,
	}, nil
}
