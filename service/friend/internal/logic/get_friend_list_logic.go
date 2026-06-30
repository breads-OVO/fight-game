package logic

import (
	"context"
	"fmt"

	"fight-game/pb/friend"
	"fight-game/pb/player"
	"fight-game/pb/player/rank"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFriendList 获取好友列表
func (l *GetFriendListLogic) GetFriendList(in *friend.GetFriendListRequest) (*friend.GetFriendListResponse, error) {
	// 查询已接受的好友
	var friends []model.Friend
	err := l.svcCtx.DB.Where("player_id = ? AND status = ?", in.PlayerId, 1).
		Find(&friends).Error
	if err != nil {
		logx.Errorf("GetFriendList query error: %v", err)
		return nil, err
	}

	// 也查询对方有我的情况（双向好友关系）
	var reverseFriends []model.Friend
	err = l.svcCtx.DB.Where("friend_id = ? AND status = ?", in.PlayerId, 1).
		Find(&reverseFriends).Error
	if err != nil {
		logx.Errorf("GetFriendList reverse query error: %v", err)
		return nil, err
	}

	// 合并好友ID集合
	friendIds := make(map[string]bool)
	for _, f := range friends {
		friendIds[f.FriendID] = true
	}
	for _, f := range reverseFriends {
		friendIds[f.PlayerID] = true
	}

	// 查询好友在线状态
	onlineSet := make(map[string]bool)
	for fid := range friendIds {
		key := fmt.Sprintf("social:online:%s", fid)
		exists, err := l.svcCtx.RedisClient.ExistsCtx(l.ctx, key)
		if err == nil && exists {
			onlineSet[fid] = true
		}
	}

	// 批量从 Player 服务获取好友详细信息（昵称、头像、等级）
	profileMap := l.fetchPlayerProfiles(friendIds)

	// 组装响应
	friendInfos := make([]*friend.FriendInfo, 0, len(friendIds))
	for fid := range friendIds {
		isOnline := onlineSet[fid]
		fp := profileMap[fid]

		fi := &friend.FriendInfo{
			PlayerId: fid,
			Status:   friend.FriendStatus_ACCEPTED,
			IsOnline: isOnline,
		}
		if fp != nil {
			fi.Nickname = fp.Nickname
			fi.AvatarUrl = fp.AvatarUrl
			fi.Level = fp.Level
			fi.Rating = fp.Rating
		}
		friendInfos = append(friendInfos, fi)
	}

	onlineCount := 0
	for _, fi := range friendInfos {
		if fi.IsOnline {
			onlineCount++
		}
	}

	return &friend.GetFriendListResponse{
		Friends:     friendInfos,
		TotalCount:  int32(len(friendInfos)),
		OnlineCount: int32(onlineCount),
	}, nil
}

// fetchPlayerProfiles 批量获取好友的昵称、头像、等级、段位分
func (l *GetFriendListLogic) fetchPlayerProfiles(friendIds map[string]bool) map[string]*friendProfile {
	result := make(map[string]*friendProfile)

	for fid := range friendIds {
		// 获取基本信息
		profileResp, err := l.svcCtx.PlayerClient.GetProfile(l.ctx, &player.GetProfileRequest{
			PlayerId: fid,
		})
		if err != nil {
			logx.Errorf("fetchPlayerProfile GetProfile error: player=%s, err=%v", fid, err)
			continue
		}
		if profileResp == nil || profileResp.Profile == nil {
			continue
		}

		p := profileResp.Profile
		fp := &friendProfile{
			Nickname:  p.Nickname,
			AvatarUrl: p.AvatarUrl,
			Level:     p.Level,
		}

		// 获取段位分
		ratingResp, err := l.svcCtx.PlayerClient.GetRating(l.ctx, &rank.GetRatingRequest{
			PlayerId: fid,
		})
		if err != nil {
			logx.Errorf("fetchPlayerProfile GetRating error: player=%s, err=%v", fid, err)
		} else if ratingResp != nil && ratingResp.Rating != nil {
			fp.Rating = int32(ratingResp.Rating.Rating)
		}

		result[fid] = fp
	}

	return result
}

// friendProfile 缓存的好友玩家信息
type friendProfile struct {
	Nickname  string
	AvatarUrl string
	Level     int32
	Rating    int32
}
