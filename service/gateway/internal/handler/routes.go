package handler

import (
	"net/http"

	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterRoutes HTTP 路由注册
func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {

	//初始化jwt中间件和websocket处理器
	jwtMiddleware := NewJwtMiddleware(svcCtx.Config.Auth.JwtSecret)
	wsHandler := NewWSHandler(svcCtx)

	/*
		注册get方法路径为ws/的路由
		当客户端发起 WebSocket 握手请求时，首先经过 jwtMiddleware.Verify，验证 JWT 的有效性。
		若验证通过，中间件将用户 ID（如 playerId）注入请求头（例如 X-Player-Id），然后调用 wsHandler.HandleWS 执行协议升级，建立 WebSocket 连接。
		若验证失败，握手请求直接返回 401 错误，不会建立 WebSocket 连接。
	*/
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws",
		Handler: jwtMiddleware.Verify(wsHandler.HandleWS),
	})

	/*
		一个简单的 GET 端点 /health，返回 JSON 格式的 {"status":"ok"}。
	*/
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		},
	})
}
