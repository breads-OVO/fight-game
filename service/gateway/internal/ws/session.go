package ws

import (
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

// SessionManager 会话管理
type SessionManager struct {
	mu          sync.RWMutex           // 读写锁
	connections map[string]*Connection // playerId -> Connection 玩家id-> 连接
}

// NewSessionManager 创建一个会话管理
func NewSessionManager() *SessionManager {
	return &SessionManager{
		connections: make(map[string]*Connection),
	}
}

// Add 添加会话
func (sm *SessionManager) Add(playerId string, conn *Connection) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if old, ok := sm.connections[playerId]; ok {
		old.Close()
	}
	sm.connections[playerId] = conn
	logx.Infof("会话添加 : player=%s, total=%d", playerId, len(sm.connections))
}

// Get 获取会话连接
func (sm *SessionManager) Get(playerId string) (*Connection, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	conn, ok := sm.connections[playerId]
	return conn, ok
}

// Remove 删除会话
func (sm *SessionManager) Remove(playerId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if conn, ok := sm.connections[playerId]; ok {
		conn.Close()
		delete(sm.connections, playerId)
		logx.Infof("删除会话: player=%s, total=%d", playerId, len(sm.connections))
	}
}

func (sm *SessionManager) Broadcast(msgType int, data []byte) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, conn := range sm.connections {
		conn.WriteMessage(msgType, data)
	}
}

func (sm *SessionManager) BroadcastToRoom(roomId string, msgType int, data []byte) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, conn := range sm.connections {
		if conn.RoomId == roomId {
			conn.WriteMessage(msgType, data)
		}
	}
}

func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.connections)
}

func (sm *SessionManager) GetAll() map[string]*Connection {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make(map[string]*Connection, len(sm.connections))
	for k, v := range sm.connections {
		result[k] = v
	}
	return result
}
