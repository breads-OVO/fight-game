package server

import (
	"context"

	"fight-game/pb/player"
	"fight-game/pb/player/asset"
	"fight-game/pb/player/currency"
	"fight-game/pb/player/rank"
	"fight-game/service/player/internal/logic"
	"fight-game/service/player/internal/svc"
)

type PlayerServiceServer struct {
	svcCtx *svc.ServiceContext
	player.UnimplementedPlayerServiceServer
}

func NewPlayerServiceServer(svcCtx *svc.ServiceContext) *PlayerServiceServer {
	return &PlayerServiceServer{
		svcCtx: svcCtx,
	}
}

// 玩家基本信息

func (s *PlayerServiceServer) GetProfile(ctx context.Context, in *player.GetProfileRequest) (*player.GetProfileResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.GetProfile(in)
}

func (s *PlayerServiceServer) LevelUp(ctx context.Context, in *player.LevelUpRequest) (*player.LevelUpResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.LevelUp(in)
}

func (s *PlayerServiceServer) ChangeNickname(ctx context.Context, in *player.UpdateNicknameRequest) (*player.UpdateNicknameResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.ChangeNickname(in)
}

func (s *PlayerServiceServer) ChangeAvatar(ctx context.Context, in *player.UpdateAvatarRequest) (*player.UpdateAvatarResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.ChangeAvatar(in)
}

func (s *PlayerServiceServer) ChangeSignature(ctx context.Context, in *player.UpdateSignatureRequest) (*player.UpdateSignatureResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.ChangeSignature(in)
}

//货币

func (s *PlayerServiceServer) GetCurrencies(ctx context.Context, in *currency.GetCurrenciesRequest) (*currency.GetCurrenciesResponse, error) {
	l := logic.NewPlayerCurrencyLogic(ctx, s.svcCtx)
	return l.GetCurrencies(in)
}

func (s *PlayerServiceServer) ChangeCurrency(ctx context.Context, in *currency.ChangeCurrencyRequest) (*currency.ChangeCurrencyResponse, error) {
	l := logic.NewPlayerCurrencyLogic(ctx, s.svcCtx)
	return l.ChangeCurrency(in)
}

//资产

func (s *PlayerServiceServer) GetInventory(ctx context.Context, in *asset.GetInventoryRequest) (*asset.GetInventoryResponse, error) {
	l := logic.NewPlayerAssetLogic(ctx, s.svcCtx)
	return l.GetInventory(in)
}

func (s *PlayerServiceServer) AddAsset(ctx context.Context, in *asset.AddAssetRequest) (*asset.AddAssetResponse, error) {
	l := logic.NewPlayerAssetLogic(ctx, s.svcCtx)
	return l.AddAsset(in)
}

func (s *PlayerServiceServer) RemoveAsset(ctx context.Context, in *asset.RemoveAssetRequest) (*asset.RemoveAssetResponse, error) {
	l := logic.NewPlayerAssetLogic(ctx, s.svcCtx)
	return l.RemoveAsset(in)
}

// 段位

func (s *PlayerServiceServer) GetRating(ctx context.Context, in *rank.GetRatingRequest) (*rank.GetRatingResponse, error) {
	l := logic.NewPlayerRankLogic(ctx, s.svcCtx)
	return l.GetRating(in)
}

func (s *PlayerServiceServer) UpdateRating(ctx context.Context, in *rank.UpdateRatingRequest) (*rank.UpdateRatingResponse, error) {
	l := logic.NewPlayerRankLogic(ctx, s.svcCtx)
	return l.UpdateRating(in)
}

func (s *PlayerServiceServer) GetMatchStats(ctx context.Context, in *rank.GetMatchStatsRequest) (*rank.GetMatchStatsResponse, error) {
	l := logic.NewPlayerRankLogic(ctx, s.svcCtx)
	return l.GetMatchStats(in)
}

func (s *PlayerServiceServer) SearchPlayer(ctx context.Context, in *player.SearchPlayerRequest) (*player.SearchPlayerResponse, error) {
	l := logic.NewPlayerLogic(ctx, s.svcCtx)
	return l.SearchPlayer(in)
}
