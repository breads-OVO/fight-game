package handler

import (
	"fight-game/pb/common"
	"fight-game/service/gateway/internal/router"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) Routes() map[common.WSMsgType]router.HandlerFunc {
	return map[common.WSMsgType]router.HandlerFunc{
		common.WSMsgType_MSG_HEARTBEAT_PING: h.Ping,
		common.WSMsgType_MSG_HEARTBEAT_PONG: h.Pong,
	}
}

func (h *SystemHandler) Ping(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	return &common.WSResponse{Code: 0, Message: "pong"}, nil
}

func (h *SystemHandler) Pong(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	return &common.WSResponse{Code: 0, Message: "pong ignored"}, nil
}
