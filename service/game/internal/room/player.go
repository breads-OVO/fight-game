package room

import (
	"sync"
	"time"

	"fight-game/pb/game"
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

	// 气槽系统
	Meter    int32 // 当前气槽值 (0~300)
	MaxMeter int32 // 最大气槽值

	// 技能状态
	SkillCooldown int32 // 技能冷却剩余帧
	SkillActive   bool  // 技能是否激活中

	// 连击系统
	ComboTimer   int32 // 连击计时器（帧数，归零则连击中断）
	ComboCount   int32 // 当前连击数
	LastHitFrame int32 // 最后一击的帧号

	// 受击状态
	HitstunRemaining int32 // 剩余硬直帧
	BlockstunRemain  int32 // 剩余防御硬直帧
	IsDown           bool  // 是否倒地

	// 帧输入
	InputMu     sync.Mutex
	LatestInput int32 // 最新帧输入（actionMask）
	InputFrame  int32 // 输入对应的帧号
	InputReady  bool  // 是否有新输入

	// 断线重连
	Connected      bool      // 是否有活跃的 WebSocket 连接
	DisconnectTime time.Time // 断线时间戳（零值表示未断线）
}

// NewPlayerData 初始化玩家数据
func NewPlayerData(playerId string) *PlayerData {
	return &PlayerData{
		PlayerId: playerId,
		MaxHP:    100,
		MaxMeter: 300,
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
	p.Meter = 0
	p.SkillCooldown = 0
	p.SkillActive = false
	p.ComboTimer = 0
	p.ComboCount = 0
	p.LastHitFrame = 0
	p.HitstunRemaining = 0
	p.BlockstunRemain = 0
	p.IsDown = false
}

// ResetMeterForNewCharacter 换角色时重置气槽
func (p *PlayerData) ResetMeterForNewCharacter() {
	p.Meter = 0
	p.SkillCooldown = 0
	p.SkillActive = false
}

// AddMeter 增加气槽值
func (p *PlayerData) AddMeter(amount int32) {
	p.Meter += amount
	if p.Meter > p.MaxMeter {
		p.Meter = p.MaxMeter
	}
}

// CanUseSkill 是否可以使用技能（检查气槽和冷却）
func (p *PlayerData) CanUseSkill() bool {
	return p.Meter >= 100 && p.SkillCooldown <= 0
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

// CurrentCharacterStats 获取当前角色属性
func (p *PlayerData) CurrentCharacterStats() *CharacterStats {
	return GetCharacterStats(p.CurrentCharacterId())
}
