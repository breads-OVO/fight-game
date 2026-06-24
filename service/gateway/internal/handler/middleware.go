package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"fight-game/pb/auth/token"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// AuthMiddleware 认证中间件（通过 Auth gRPC 验证 Token）
type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

// Verify 通过 AuthRPC.VerifyToken 验证 JWT，将 playerId 注入请求头
func (m *AuthMiddleware) Verify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			httpx.Error(w, errors.New("missing authorization header"))
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		resp, err := m.svcCtx.AuthClient.VerifyToken(context.Background(), &token.VerifyRequest{
			Token: tokenStr,
		})
		if err != nil || !resp.Valid {
			logx.Errorf("auth verify token failed: %v", err)
			httpx.Error(w, errors.New("invalid token"))
			return
		}

		r.Header.Set("X-Player-Id", resp.PlayerId)
		next(w, r)
	}
}
