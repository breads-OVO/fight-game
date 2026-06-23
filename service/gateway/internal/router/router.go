package router

import (
	"fight-game/pb/common"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type HandlerFunc func(playerId string, msg *common.WSMessage) (*common.WSResponse, error)

type Router struct {
	handlers map[common.WSMsgType]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[common.WSMsgType]HandlerFunc),
	}
}

// Register 注册
func (r *Router) Register(msgType common.WSMsgType, handler HandlerFunc) {
	r.handlers[msgType] = handler
}

// RegisterModule 批量注册
func (r *Router) RegisterModule(handlers map[common.WSMsgType]HandlerFunc) {
	for msgType, handler := range handlers {
		r.handlers[msgType] = handler
	}
}

// Dispatch 分发
func (r *Router) Dispatch(playerId string, data []byte) (*common.WSResponse, error) {
	var msg common.WSMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		logx.Errorf("proto unmarshal error: %v", err)
		return &common.WSResponse{Code: -1, Message: "invalid protobuf message"}, err
	}
	msg.PlayerId = playerId

	handler, ok := r.handlers[common.WSMsgType(msg.MsgType)]
	if !ok {
		logx.Infof("no handler for msgType=%d", msg.MsgType)
		return &common.WSResponse{Code: -2, Message: "unknown msg type"}, nil
	}

	return handler(playerId, &msg)
}
