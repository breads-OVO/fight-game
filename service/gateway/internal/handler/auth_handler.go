package handler

import (
	"context"
	"io"
	"net/http"

	"fight-game/pb/auth/login"
	"fight-game/pb/auth/register"
	"fight-game/pb/auth/token"
	"fight-game/pb/common"
	"fight-game/pkg/common/utils"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type AuthHandler struct {
	svcCtx *svc.ServiceContext
}

func NewAuthHandler(svcCtx *svc.ServiceContext) *AuthHandler {
	return &AuthHandler{svcCtx: svcCtx}
}

// Routes 返回 WS 消息路由（已认证的令牌刷新）
func (h *AuthHandler) Routes() map[common.WSMsgType]router.HandlerFunc {
	return map[common.WSMsgType]router.HandlerFunc{
		common.WSMsgType_MSG_AUTH_TOKEN_REFRESH: h.RefreshToken,
	}
}

// RefreshToken 处理 WS 令牌刷新请求
func (h *AuthHandler) RefreshToken(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req token.RefreshRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}

	resp, err := h.svcCtx.AuthClient.RefreshToken(context.Background(), &req)
	if err != nil {
		logx.Errorf("auth refresh token failed: %v", err)
		return &common.WSResponse{Code: -1, Message: "refresh token failed"}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// --- HTTP 端点 ---

// HandleLogin 处理登录请求
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req login.LoginRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 调用 AuthRPC
	resp, err := h.svcCtx.AuthClient.Login(context.Background(), &req)
	if err != nil {
		logx.Errorf("auth login failed: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(data)
}

// HandleRegister 处理注册请求（Body: protobuf RegisterRequest）
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req register.RegisterRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.svcCtx.AuthClient.Register(context.Background(), &req)
	if err != nil {
		logx.Errorf("auth register failed: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(data)
}

// HandleRefresh 处理令牌刷新请求（Body: protobuf RefreshRequest）
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req token.RefreshRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.svcCtx.AuthClient.RefreshToken(context.Background(), &req)
	if err != nil {
		logx.Errorf("auth refresh token failed: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	data, _ := proto.Marshal(resp)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(data)
}
