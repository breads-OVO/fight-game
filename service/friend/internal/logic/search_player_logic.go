package logic

import (
	"context"

	"fight-game/pb/friend"
	"fight-game/pb/player"
	"fight-game/pb/player/rank"
	"fight-game/service/friend/internal/model"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchPlayerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchPlayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchPlayerLogic {
	return &SearchPlayerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchPlayer 搜索玩家（通过 Player 服务按昵称/ID 模糊搜索）
func (l *SearchPlayerLogic) SearchPlayer(in *friend.SearchPlayerRequest) (*friend.SearchPlayerResponse, error) {
	// 调用 Player 服务搜索
	searchResp, err := l.svcCtx.PlayerClient.SearchPlayer(l.ctx, &player.SearchPlayerRequest{
		Keyword:  in.Keyword,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		logx.Errorf("SearchPlayer call PlayerRpc error: keyword=%s, err=%v", in.Keyword, err)
		return &friend.SearchPlayerResponse{
			Players: []*friend.PlayerSearchResult{},
			Total:   0,
		}, nil
	}

	// 查询当前玩家的好友列表，标记 isFriend
	myFriendIds := l.getFriendIds(in.PlayerId)

	// 转换结果
	results := make([]*friend.PlayerSearchResult, 0, len(searchResp.Players))
	for _, p := range searchResp.Players {
		// 获取段位分
		rating := int32(0)
		ratingResp, err := l.svcCtx.PlayerClient.GetRating(l.ctx, &rank.GetRatingRequest{
			PlayerId: p.PlayerId,
		})
		if err == nil && ratingResp != nil && ratingResp.Rating != nil {
			rating = int32(ratingResp.Rating.Rating)
		}

		results = append(results, &friend.PlayerSearchResult{
			PlayerId:  p.PlayerId,
			Nickname:  p.Nickname,
			AvatarUrl: p.AvatarUrl,
			Level:     p.Level,
			Rating:    rating,
			IsFriend:  myFriendIds[p.PlayerId],
		})
	}

	return &friend.SearchPlayerResponse{
		Players: results,
		Total:   searchResp.Total,
	}, nil
}

// getFriendIds 获取玩家所有好友ID集合
func (l *SearchPlayerLogic) getFriendIds(playerId string) map[string]bool {
	result := make(map[string]bool)

	var friends []model.Friend
	if err := l.svcCtx.DB.Where("player_id = ? AND status = ?", playerId, 1).
		Find(&friends).Error; err != nil {
		return result
	}

	for _, f := range friends {
		result[f.FriendID] = true
	}

	// 也查反向
	var reverseFriends []model.Friend
	if err := l.svcCtx.DB.Where("friend_id = ? AND status = ?", playerId, 1).
		Find(&reverseFriends).Error; err != nil {
		return result
	}
	for _, f := range reverseFriends {
		result[f.PlayerID] = true
	}

	return result
}
