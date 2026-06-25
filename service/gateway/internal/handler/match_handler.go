package handler

import (
	"context"

	"fight-game/pb/common"
	"fight-game/pb/match/queue"
	"fight-game/pkg/common/utils"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type MatchHandler struct {
	svcCtx *svc.ServiceContext
}

func NewMatchHandler(svcCtx *svc.ServiceContext) *MatchHandler {
	return &MatchHandler{svcCtx: svcCtx}
}

// Routes 返回匹配相关的 WS 消息路由
func (h *MatchHandler) Routes() map[common.WSMsgType]router.HandlerFunc {
	return map[common.WSMsgType]router.HandlerFunc{
		common.WSMsgType_MSG_MATCH_QUEUE:       h.JoinQueue,
		common.WSMsgType_MSG_MATCH_QUEUE_LEAVE: h.LeaveQueue,
		common.WSMsgType_MSG_MATCH_STATUS:      h.GetMatchStatus,
	}
}

// JoinQueue 处理加入匹配队列请求
func (h *MatchHandler) JoinQueue(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req queue.MatchRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MatchClient.JoinQueue(context.Background(), &req)
	if err != nil {
		logx.Errorf("match join queue failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// LeaveQueue 处理离开匹配队列请求
func (h *MatchHandler) LeaveQueue(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req queue.LeaveQueueRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MatchClient.LeaveQueue(context.Background(), &req)
	if err != nil {
		logx.Errorf("match leave queue failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetMatchStatus 处理查询匹配状态请求
func (h *MatchHandler) GetMatchStatus(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req queue.MatchStatusRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}

	resp, err := h.svcCtx.MatchClient.GetMatchStatus(context.Background(), &req)
	if err != nil {
		logx.Errorf("match get status failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}
