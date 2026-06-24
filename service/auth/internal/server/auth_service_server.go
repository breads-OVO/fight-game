package server

import (
	"context"

	"fight-game/pb/auth"
	"fight-game/pb/auth/login"
	"fight-game/pb/auth/register"
	"fight-game/pb/auth/token"
	"fight-game/service/auth/internal/logic"
	"fight-game/service/auth/internal/svc"
)

type AuthServiceServer struct {
	svcCtx *svc.ServiceContext
	auth.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(svcCtx *svc.ServiceContext) *AuthServiceServer {
	return &AuthServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AuthServiceServer) Login(ctx context.Context, in *login.LoginRequest) (*login.LoginResponse, error) {
	l := logic.NewLoginLogic(ctx, s.svcCtx)
	return l.Login(in)
}

func (s *AuthServiceServer) Register(ctx context.Context, in *register.RegisterRequest) (*register.RegisterResponse, error) {
	l := logic.NewRegisterLogic(ctx, s.svcCtx)
	return l.Register(in)
}

func (s *AuthServiceServer) VerifyToken(ctx context.Context, in *token.VerifyRequest) (*token.VerifyResponse, error) {
	l := logic.NewTokenLogic(ctx, s.svcCtx)
	return l.VerifyToken(in)
}

func (s *AuthServiceServer) RefreshToken(ctx context.Context, in *token.RefreshRequest) (*token.RefreshResponse, error) {
	l := logic.NewTokenLogic(ctx, s.svcCtx)
	return l.RefreshToken(in)
}
