package logic

import (
	"context"
	"errors"
	"fight-game/pb/player/rank"
	"fight-game/service/player/internal/model"
	"fight-game/service/player/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PlayerRankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlayerRankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerRankLogic {
	return &PlayerRankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetRating 获取段位分
func (l *PlayerRankLogic) GetRating(in *rank.GetRatingRequest) (*rank.GetRatingResponse, error) {
	playerId := in.GetPlayerId()
	season := in.GetSeason()

	var playerRank model.PlayerRank
	result := l.svcCtx.DB.Where("player_id = ? AND season = ?", playerId, season).First(&playerRank)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player rank not found for player %s in season %s", playerId, season)
		}
		return nil, status.Errorf(codes.Internal, "failed to query player rank: %v", result.Error)
	}

	return &rank.GetRatingResponse{
		Rating: &rank.RatingInfo{
			PlayerId: playerId,
			Rating:   playerRank.Rating,
		},
	}, nil

}

// UpdateRating 修改段位分
func (l *PlayerRankLogic) UpdateRating(in *rank.UpdateRatingRequest) (*rank.UpdateRatingResponse, error) {
	playerId := in.GetPlayerId()
	delta := in.GetDelta()
	season := l.svcCtx.Config.Player.Season

	//查询，有则更新，无则插入
	var playerRank model.PlayerRank
	result := l.svcCtx.DB.Where("player_id = ? AND season = ?", playerId, season).First(&playerRank)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 插入
			playerRank := model.PlayerRank{
				PlayerId: playerId,
				Season:   season,
				Rating:   l.svcCtx.Config.Player.RatingDefault + delta,
			}
			result := l.svcCtx.DB.Create(&playerRank)
			if result.Error != nil {
				return nil, status.Errorf(codes.Internal, "failed to create player rank: %v", result.Error)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "failed to query player rank: %v", result.Error)
		}
	} else {
		// 更新
		playerRank.Rating += delta
		result := l.svcCtx.DB.Save(&playerRank)
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "failed to update player rank: %v", result.Error)
		}
	}
	return &rank.UpdateRatingResponse{
		Rating: &rank.RatingInfo{
			PlayerId: playerId,
			Rating:   playerRank.Rating,
		},
	}, nil
}

// GetMatchStats 获取段位信息
func (l *PlayerRankLogic) GetMatchStats(in *rank.GetMatchStatsRequest) (*rank.GetMatchStatsResponse, error) {
	playerId := in.GetPlayerId()
	season := l.svcCtx.Config.Player.Season

	var playerMatchStats model.PlayerRank
	result := l.svcCtx.DB.Where("player_id = ? AND season = ?", playerId, season).First(&playerMatchStats)
	winRate := 0.00
	//为空则用默认值
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			playerMatchStats = model.PlayerRank{
				PlayerId:  playerId,
				Season:    season,
				Rating:    l.svcCtx.Config.Player.RatingDefault,
				TotalGame: 0,
				Win:       0,
				Lose:      0,
			}
		} else {
			return nil, status.Errorf(codes.Internal, "failed to query player rank: %v", result.Error)
		}
	} else {
		if playerMatchStats.TotalGame > 0 {
			win := playerMatchStats.Win
			winRate = float64(win / playerMatchStats.TotalGame)
		}
	}

	return &rank.GetMatchStatsResponse{
		Stats: &rank.MatchStatsInfo{
			TotalGames: playerMatchStats.TotalGame,
			Wins:       playerMatchStats.Win,
			Loses:      playerMatchStats.Lose,
			WinRate:    winRate,
			Season:     season,
		},
	}, nil

}
