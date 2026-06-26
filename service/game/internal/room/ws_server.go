package room

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
	ReadBufferSize:  256,
	WriteBufferSize: 1024,
}

// GameConn 游戏连接
type GameConn struct {
	PlayerId string
	RoomId   string
	conn     *websocket.Conn
	mu       sync.Mutex
	closed   bool
}

// NewGameConn 创建游戏连接
func NewGameConn(playerId, roomId string, ws *websocket.Conn) *GameConn {
	return &GameConn{
		PlayerId: playerId,
		RoomId:   roomId,
		conn:     ws,
	}
}

// Send 发送消息
func (gc *GameConn) Send(data []byte) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if gc.closed {
		return fmt.Errorf("连接已关闭")
	}
	return gc.conn.WriteMessage(websocket.BinaryMessage, data)
}

// ReadLoop 读取循环
func (gc *GameConn) ReadLoop(room *Room) {
	defer func() {
		gc.mu.Lock()
		gc.closed = true
		gc.conn.Close()
		gc.mu.Unlock()
		room.UnregisterConn(gc.PlayerId)
	}()

	for {
		_, data, err := gc.conn.ReadMessage()
		if err != nil {
			return
		}
		// 将原始消息交给 Room 处理
		room.HandleWSMessage(gc.PlayerId, data)
	}
}

// Close 关闭连接
func (gc *GameConn) Close() {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if gc.closed {
		return
	}
	gc.closed = true
	gc.conn.Close()
}

// RoomManager 房间管理器（WS用）
type RoomManager interface {
	GetRoom(roomId string) *Room
}

// WSServer WebSocket 直连服务器
type WSServer struct {
	addr    string
	mux     *http.ServeMux
	server  *http.Server
	manager RoomManager
}

// NewWSServer 创建WS服务器
func NewWSServer(addr string, manager RoomManager) *WSServer {
	mux := http.NewServeMux()
	ws := &WSServer{
		addr:    addr,
		mux:     mux,
		manager: manager,
	}
	mux.HandleFunc("/play", ws.handlePlay)
	return ws
}

// Start 启动WS服务器
func (ws *WSServer) Start() error {
	ws.server = &http.Server{
		Addr:    ws.addr,
		Handler: ws.mux,
	}
	return ws.server.ListenAndServe()
}

// Stop 停止WS服务器
func (ws *WSServer) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}

// handlePlay 处理游戏连接请求
// 客户端连接: ws://gameAddr/play?roomId=xxx&playerId=xxx
func (ws *WSServer) handlePlay(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Query().Get("roomId")
	playerId := r.URL.Query().Get("playerId")

	if roomId == "" || playerId == "" {
		http.Error(w, "roomId and playerId required", http.StatusBadRequest)
		return
	}

	room := ws.manager.GetRoom(roomId)
	if room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	gc := NewGameConn(playerId, roomId, conn)
	room.RegisterConn(playerId, gc)

	gc.ReadLoop(room)
}
