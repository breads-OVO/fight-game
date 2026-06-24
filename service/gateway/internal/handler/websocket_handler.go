package handler

import (
	"errors"
	"fight-game/pb/common"
	"fight-game/pkg/common/utils"
	"fight-game/service/gateway/internal/config"
	"net/http"
	"time"

	"fight-game/service/gateway/internal/svc"
	"fight-game/service/gateway/internal/ws"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// websocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler websocket处理器
type WSHandler struct {
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewWSHandler 创建websocket处理器
func NewWSHandler(svcCtx *svc.ServiceContext) *WSHandler {
	return &WSHandler{svcCtx: svcCtx}
}

// HandleWS 处理websocket请求
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	playerId := r.Header.Get("X-Player-Id")
	if playerId == "" {
		httpx.Error(w, errors.New("unauthorized"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("websocket upgrade failed: %v", err)
		return
	}

	wrapper := ws.NewConnection(playerId, conn)
	h.svcCtx.SessionMgr.Add(playerId, wrapper)

	logx.Infof("websocket 连接: player=%s, addr=%s", playerId, r.RemoteAddr)

	cfg := h.svcCtx.Config.WebSocket
	go h.readLoop(wrapper, cfg)
}

func (h *WSHandler) readLoop(conn *ws.Connection, cfg config.WebSocket) {
	defer func() {
		h.svcCtx.SessionMgr.Remove(conn.PlayerId)
	}()

	// 设置读取消息缓冲区最大字节，超时时间，Ping响应
	conn.Conn.SetReadLimit(cfg.MaxMessageSize)
	err := conn.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.IdleTimeout) * time.Second))
	if err != nil {
		logx.Errorf("websocket 设置超时时间错误: player=%s, err=%v", conn.PlayerId, err)
		return
	}
	conn.Conn.SetPongHandler(func(string) error {
		conn.MarkActive()
		err := conn.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.IdleTimeout) * time.Second))
		if err != nil {
			logx.Errorf("websocket 设置Ping响应错误: player=%s, err=%v", conn.PlayerId, err)
			return err
		}
		return nil
	})

	// 读取消息循环
	for {
		msgType, data, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logx.Errorf("websocket 读取错误: player=%s, err=%v", conn.PlayerId, err)
			}
			break
		}

		conn.MarkActive()

		// 处理Ping消息
		if msgType == websocket.PingMessage {
			err := conn.Conn.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				logx.Errorf("websocket 响应Ping错误: player=%s, err=%v", conn.PlayerId, err)
				return
			}
			continue
		}

		// 只处理二进制消息（Protobuf）
		if msgType == websocket.BinaryMessage {
			msg, err := utils.UnpackWSMessage(data)
			if err != nil {
				logx.Errorf("proto unmarshal error: %v", err)
				return
			}
			resp, err := h.svcCtx.Router.Dispatch(conn.PlayerId, msg)
			if err != nil {
				logx.Errorf("route message error: player=%s, err=%v", conn.PlayerId, err)
				resp = &common.WSResponse{Code: -1, Message: "internal error"}
			}
			if resp != nil {
				err := conn.WriteProtobuf(0, resp)
				if err != nil {
					logx.Errorf("websocket 发送消息错误: player=%s, err=%v", conn.PlayerId, err)
					return
				}
			}
		}
	}
}
