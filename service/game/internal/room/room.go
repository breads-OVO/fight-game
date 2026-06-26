package room

import (
	"fmt"
	"math"
	"sync"
	"time"

	"fight-game/pb/common"
	"fight-game/pb/game"
	pbQueue "fight-game/pb/match/queue"

	"google.golang.org/protobuf/proto"
)

const (
	disconnectTimeout = 5 * time.Second // 断线判负超时
)

// RoomConfig 房间配置
type RoomConfig struct {
	PickTimeout  int    // 选人超时（秒）
	BanTimeout   int    // 禁用超时（秒）
	FightTimeout int    // 每回合战斗超时（秒）
	Characters   int    // 每人可选角色数
	WsAddr       string // Game 直连地址
}

// Room 游戏房间
type Room struct {
	RoomId   string           // 房间ID
	GameType pbQueue.GameType // 游戏类型
	Players  []*PlayerData    // 玩家数据
	WsAddr   string           // Game 直连地址

	cfg     RoomConfig      // 配置
	stage   game.GameStage  // 当前阶段
	stageCh chan StageState // 阶段变更通知
	closeCh chan struct{}   // 关闭通知
	mu      sync.RWMutex

	// Pick 阶段状态
	pickReady map[string]bool // playerId → 已确认

	// Ban 阶段状态（竞技）
	banCount     int             // 禁用数
	banCharacter []string        // 禁用的角色
	banConfirmed map[string]bool // playerId → 已确认

	// Fight 阶段状态
	RoundNo      int32        // 当前回合（从1开始）
	FrameNo      int32        // 当前帧号
	RoundWinner  string       // 当前回合胜者
	RoundTimeout bool         // 当前回合是否超时
	RoundEnded   bool         // 当前回合是否结束
	fightTick    *time.Ticker // 战斗计时器

	// 结果
	RoundResults []*game.RoundResult // 回合结果
	GameResult   *game.GameResult    // 游戏结果

	// 连接管理
	conns   map[string]*GameConn // playerId → WS连接
	connsMu sync.RWMutex

	// 房间清理回调（结算后自动调用）
	onCleanup func()
}

// StageState 阶段状态（通过 channel 通知外部）
type StageState struct {
	Stage     game.GameStage
	Countdown int
	Data      interface{} // 附加数据
}

// NewRoom 创建新房间
func NewRoom(roomId string, gameType pbQueue.GameType, playerIds []string, rating int32, cfg RoomConfig) *Room {
	players := make([]*PlayerData, len(playerIds))
	for i, pid := range playerIds {
		players[i] = NewPlayerData(pid)
	}

	r := &Room{
		RoomId:       roomId,
		GameType:     gameType,
		Players:      players,
		WsAddr:       cfg.WsAddr,
		cfg:          cfg,
		stage:        game.GameStage_STAGE_UNKNOWN,
		stageCh:      make(chan StageState, 8),
		closeCh:      make(chan struct{}),
		pickReady:    make(map[string]bool),
		conns:        make(map[string]*GameConn),
		RoundResults: make([]*game.RoundResult, 0),
	}
	if gameType == pbQueue.GameType_COMPETITION {
		//竞技模式ban位
		r.banCount = getBanCount(rating)
	}
	return r
}

// PlayerIdList 获取所有玩家ID
func (r *Room) PlayerIdList() []string {
	ids := make([]string, len(r.Players))
	for i, p := range r.Players {
		ids[i] = p.PlayerId
	}
	return ids
}

// Start 启动房间阶段循环
func (r *Room) Start() {
	if r.GameType == 1 {
		r.startBanStage()
	} else {
		r.startPickStage()
	}
}

// Stop 停止房间
func (r *Room) Stop() {
	close(r.closeCh)
}

// 获取对手
func (r *Room) opponentOf(playerId string) *PlayerData {
	for _, p := range r.Players {
		if p.PlayerId != playerId {
			return p
		}
	}
	return nil
}

// broadcast 向所有玩家广播消息
func (r *Room) broadcast(data []byte) {
	r.connsMu.RLock()
	defer r.connsMu.RUnlock()
	for _, conn := range r.conns {
		conn.Send(data)
	}
}

// broadcastStageState 广播阶段状态
func (r *Room) broadcastStageState(stage game.GameStage, countdown int) {
	state := &game.StageState{
		Stage:     stage,
		Countdown: int32(countdown),
	}
	data, _ := proto.Marshal(state)
	if data != nil {
		r.broadcast(data)
	}
}

// RegisterConn 注册玩家WS连接
func (r *Room) RegisterConn(playerId string, conn *GameConn) {
	r.connsMu.Lock()
	// 关闭旧连接（如果存在）
	if old, ok := r.conns[playerId]; ok {
		old.Close()
	}
	r.conns[playerId] = conn
	r.connsMu.Unlock()
}

// UnregisterConn 注销玩家WS连接（只有同一连接对象才删除，防止重连竞争）
func (r *Room) UnregisterConn(playerId string, conn *GameConn) {
	r.connsMu.Lock()
	if r.conns[playerId] == conn {
		delete(r.conns, playerId)
	}
	r.connsMu.Unlock()
}

// GetPlayer 获取玩家数据
func (r *Room) GetPlayer(playerId string) *PlayerData {
	for _, p := range r.Players {
		if p.PlayerId == playerId {
			return p
		}
	}
	return nil
}

// HasPlayer 检查玩家是否属于该房间
func (r *Room) HasPlayer(playerId string) bool {
	for _, p := range r.Players {
		if p.PlayerId == playerId {
			return true
		}
	}
	return false
}

// MarkDisconnected 标记玩家断线
func (r *Room) MarkDisconnected(playerId string) {
	for _, p := range r.Players {
		if p.PlayerId == playerId {
			p.Connected = false
			p.DisconnectTime = time.Now()
			r.log("玩家 %s 断线", playerId)
			return
		}
	}
}

// MarkReconnected 标记玩家重连
func (r *Room) MarkReconnected(playerId string) {
	for _, p := range r.Players {
		if p.PlayerId == playerId {
			p.Connected = true
			p.DisconnectTime = time.Time{}
			r.log("玩家 %s 重连成功", playerId)
			return
		}
	}
}

// SetCleanup 设置房间清理回调（结算后自动调用）
func (r *Room) SetCleanup(fn func()) {
	r.mu.Lock()
	r.onCleanup = fn
	r.mu.Unlock()
}

// log 房间日志
func (r *Room) log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[Room %s] %s\n", r.RoomId, msg)
}

// HandleWSMessage 处理从WS连接收到的消息
func (r *Room) HandleWSMessage(playerId string, data []byte) {
	// 解析 WSMessage 包装
	var msg common.WSMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		r.log("消息解析失败: %v", err)
		return
	}

	switch common.WSMsgType(msg.MsgType) {
	case common.WSMsgType_MSG_GAME_PICK:
		var pick game.PickRequest
		if err := proto.Unmarshal(msg.Body, &pick); err == nil {
			r.HandlePick(pick.PlayerId, pick.GetCharacters())
		}
	case common.WSMsgType_MSG_GAME_BAN:
		var ban game.BanRequest
		if err := proto.Unmarshal(msg.Body, &ban); err == nil {
			r.HandleBan(ban.PlayerId, ban.CharacterIds)
		}
	case common.WSMsgType_MSG_GAME_INPUT:
		var input game.PlayerInput
		if err := proto.Unmarshal(msg.Body, &input); err == nil {
			player := r.GetPlayer(playerId)
			if player != nil {
				player.SubmitInput(input.FrameNo, input.ActionMask)
			}
		}
	case common.WSMsgType_MSG_GAME_READY:
		r.log("玩家 %s 确认准备", playerId)
	}
}

// getBanCount 根据段位分获取ban位
func getBanCount(rating int32) int {
	if rating <= 1500 {
		return 0
	}
	return int(math.Floor(float64(rating-1500) / 200))
}
