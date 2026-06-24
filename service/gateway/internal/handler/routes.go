package handler

import (
	"net/http"

	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterRoutes HTTP 路由注册
func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {

	// 初始化认证中间件（通过 AuthRPC 验证 Token）
	authMiddleware := NewAuthMiddleware(svcCtx)
	wsHandler := NewWSHandler(svcCtx)
	authHandler := NewAuthHandler(svcCtx)

	/*
		注册 WebSocket 路由
		客户端先通过 Auth HTTP 端点获取 JWT，再使用该 JWT 建立 WebSocket 连接。
		authMiddleware.Verify 通过 AuthRPC.VerifyToken 验证 JWT，将 playerId 注入请求头，
		然后调用 wsHandler.HandleWS 执行协议升级，建立 WebSocket 连接。
	*/
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws",
		Handler: authMiddleware.Verify(wsHandler.HandleWS),
	})

	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		},
	})

	// Auth HTTP 端点（无需 JWT，用于首次获取 token）
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/auth/login",
		Handler: authHandler.HandleLogin,
	})
	// 注册
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/auth/register",
		Handler: authHandler.HandleRegister,
	})
	// 刷新 token
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/auth/refresh",
		Handler: authHandler.HandleRefresh,
	})
}
