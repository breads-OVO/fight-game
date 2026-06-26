package room

import (
	"fight-game/pb/game"
	"sync"
	"time"
)

// PlayerData 玩家在房间中的数据
type PlayerData struct {
	PlayerId string

	// Ban 阶段（竞技）
	BannedConfirmed  bool     // 是否已确认
	BannedCharacters []string // 已禁用的角色ID

	// Pick 阶段
	PickedCharacters *game.CharacterConfig // 选择的角色配置
	PickedConfirmed  bool                  // 是否已确认

	// Fight 阶段
	RemainingCharacters []string // 剩余可用角色（按顺序）
	CurrentCharacterIdx int      // 当前出战角色索引
	CurrentHP           int32    // 当前角色血量
	MaxHP               int32    // 角色最大血量

	// 帧输入
	InputMu     sync.Mutex
	LatestInput int32 // 最新帧输入（actionMask）
	InputFrame  int32 // 输入对应的帧号
	InputReady  bool  // 是否有新输入

	// 断线重连
	Connected      bool      // 是否有活跃的 WebSocket 连接
	DisconnectTime time.Time // 断线时间戳（零值表示未断线）
}

const defaultMaxHP = 100

// NewPlayerData 初始化玩家数据
func NewPlayerData(playerId string) *PlayerData {
	return &PlayerData{
		PlayerId: playerId,
		MaxHP:    defaultMaxHP,
	}
}

// SubmitInput 提交帧输入
func (p *PlayerData) SubmitInput(frameNo int32, actionMask int32) {
	p.InputMu.Lock()
	defer p.InputMu.Unlock()
	p.LatestInput = actionMask
	p.InputFrame = frameNo
	p.InputReady = true
}

// ConsumeInput 消费帧输入
func (p *PlayerData) ConsumeInput() (int32, int32, bool) {
	p.InputMu.Lock()
	defer p.InputMu.Unlock()
	if !p.InputReady {
		return 0, 0, false
	}
	p.InputReady = false
	return p.LatestInput, p.InputFrame, true
}

// ResetForNewRound 新回合重置玩家战斗数据
func (p *PlayerData) ResetForNewRound(maxHP int32) {
	p.CurrentHP = maxHP
	p.MaxHP = maxHP
}

// NextCharacter 切换到下一个人物
// 返回是否还有可用角色
func (p *PlayerData) NextCharacter() bool {
	p.CurrentCharacterIdx++
	return p.CurrentCharacterIdx < len(p.RemainingCharacters)
}

// CurrentCharacterId 获取当前角色ID
func (p *PlayerData) CurrentCharacterId() string {
	if p.CurrentCharacterIdx < len(p.RemainingCharacters) {
		return p.RemainingCharacters[p.CurrentCharacterIdx]
	}
	return ""
}
