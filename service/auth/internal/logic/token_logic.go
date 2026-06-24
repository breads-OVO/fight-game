package logic

import (
	"context"

	"fight-game/pb/auth/token"
	"fight-game/pkg/common/utils"
	"fight-game/service/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TokenLogic {
	return &TokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VerifyToken 验证token
func (l *TokenLogic) VerifyToken(in *token.VerifyRequest) (*token.VerifyResponse, error) {
	playerId, err := utils.ParseToken(l.svcCtx.Config.JwtSecret, in.Token)
	if err != nil {
		return &token.VerifyResponse{Valid: false}, nil
	}

	return &token.VerifyResponse{
		Valid:    true,
		PlayerId: playerId,
	}, nil
}

// RefreshToken 刷新token
func (l *TokenLogic) RefreshToken(in *token.RefreshRequest) (*token.RefreshResponse, error) {
	playerId, err := utils.ParseToken(l.svcCtx.Config.JwtSecret, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenStr, _, err := utils.GenerateTokenPair(l.svcCtx.Config.JwtSecret, playerId)
	if err != nil {
		return nil, err
	}

	return &token.RefreshResponse{
		Token:      tokenStr,
		ExpireTime: 0,
	}, nil
}
