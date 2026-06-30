package logic

import (
	"context"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveFriendLogic {
	return &RemoveFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RemoveFriend 删除好友
func (l *RemoveFriendLogic) RemoveFriend(in *friend.RemoveFriendRequest) (*friend.RemoveFriendResponse, error) {
	// 删除双向好友关系
	result := l.svcCtx.DB.Where(
		"(player_id = ? AND friend_id = ?) OR (player_id = ? AND friend_id = ?)",
		in.PlayerId, in.FriendId, in.FriendId, in.PlayerId,
	).Delete(&model.Friend{})

	if result.Error != nil {
		logx.Errorf("RemoveFriend delete error: %v", result.Error)
		return &friend.RemoveFriendResponse{Success: false}, result.Error
	}

	logx.Infof("Friend removed: player=%s, friend=%s", in.PlayerId, in.FriendId)
	return &friend.RemoveFriendResponse{
		Success: result.RowsAffected > 0,
	}, nil
}
