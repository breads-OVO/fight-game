package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Connection struct {
	PlayerId   string          // 玩家ID
	RoomId     string          // 房间ID
	Conn       *websocket.Conn // 连接
	mu         sync.Mutex      // 互斥锁
	LastActive time.Time       // 最后活跃时间
	closed     bool            // 是否关闭
}

// NewConnection 创建一个连接
func NewConnection(playerId string, conn *websocket.Conn) *Connection {
	return &Connection{
		PlayerId:   playerId,
		Conn:       conn,
		LastActive: time.Now(),
	}
}

// WriteMessage 写入消息
func (c *Connection) WriteMessage(msgType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.LastActive = time.Now()
	return c.Conn.WriteMessage(msgType, data)
}

// WriteJSON 写入JSON消息
func (c *Connection) WriteJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, data)
}

// Close 关闭连接
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	c.Conn.Close()
	logx.Infof("connection closed: player=%s", c.PlayerId)
}

// IsClosed 检查连接是否关闭
func (c *Connection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// MarkActive 标记连接活跃
func (c *Connection) MarkActive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActive = time.Now()
}
