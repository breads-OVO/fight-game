package svc

import (
	pbQueue "fight-game/pb/match/queue"
	"sync"

	"fight-game/service/game/internal/config"
	"fight-game/service/game/internal/room"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  redis.UniversalClient

	// 房间管理
	rooms     map[string]*room.Room
	roomsMu   sync.RWMutex
	roomCount int32
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		rooms:  make(map[string]*room.Room),
	}
}

// CreateAndStartRoom 创建并启动游戏房间
func (sc *ServiceContext) CreateAndStartRoom(roomId string, gameType pbQueue.GameType, playerIds []string, raring int32) *room.Room {
	r := room.NewRoom(roomId, gameType, playerIds, raring, room.RoomConfig{
		PickTimeout:  sc.Config.Game.PickTimeout,
		BanTimeout:   sc.Config.Game.BanTimeout,
		FightTimeout: sc.Config.Game.FightTimeout,
		Characters:   sc.Config.Game.Characters,
		WsAddr:       sc.Config.Game.WsAddr,
	})

	sc.roomsMu.Lock()
	sc.rooms[roomId] = r
	sc.roomCount++
	sc.roomsMu.Unlock()

	return r
}

// GetRoom 获取房间
func (sc *ServiceContext) GetRoom(roomId string) *room.Room {
	sc.roomsMu.RLock()
	defer sc.roomsMu.RUnlock()
	return sc.rooms[roomId]
}

// RemoveRoom 移除房间
func (sc *ServiceContext) RemoveRoom(roomId string) {
	sc.roomsMu.Lock()
	delete(sc.rooms, roomId)
	sc.roomsMu.Unlock()
}
