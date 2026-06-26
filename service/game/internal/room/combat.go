package room

import (
	"math"
	"sync"
	"time"

	"fight-game/pb/game"

	"google.golang.org/protobuf/proto"
)

const (
	frameInterval = 16 * time.Millisecond // 60fps
	fps           = 60

	// 按键掩码
	ActionMaskLeft  = 1 << 0
	ActionMaskRight = 1 << 1
	ActionMaskJump  = 1 << 2
	ActionMaskLight = 1 << 3
	ActionMaskHeavy = 1 << 4
	ActionMaskGuard = 1 << 5
	ActionMaskEvade = 1 << 6
	ActionMaskSkill = 1 << 7

	// 动作状态
	ActionIdle        = 0
	ActionWalk        = 1
	ActionJump        = 2
	ActionAttackLight = 3
	ActionAttackHeavy = 4
	ActionGuard       = 5
	ActionEvade       = 6
	ActionSkill       = 7
	ActionHit         = 8
	ActionDown        = 9
	ActionBlockstun   = 10

	// 角色属性
	moveSpeed    = 3.0
	jumpVelocity = -10.0
	gravity      = 0.5
	groundY      = 400.0
	minX         = 30.0
	maxX         = 770.0

	// 攻击属性
	lightDamage  = 8
	heavyDamage  = 20
	lightRange   = 80.0
	heavyRange   = 100.0
	lightHitstun = 10 // 帧
	heavyHitstun = 20
	lightStartup = 5 // 前摇
	heavyStartup = 12
	lightActive  = 3 // 有效帧
	heavyActive  = 5

	// 防御属性
	guardDamageReduce = 0.3
	guardKnockback    = 30.0

	// 闪避
	evadeDuration = 12
	evadeSpeed    = 8.0
)

// Entity 战场实体
type Entity struct {
	PlayerId     string  // 玩家ID
	CharacterId  string  // 角色ID
	X, Y         float64 // 坐标
	Vx, Vy       float64 // 速度
	HP           int32   // 当前生命值
	MaxHP        int32   // 最大生命值
	ActionState  int32   // 动作状态
	FrameCounter int32   // 当前帧
	Facing       float64 // 1=右, -1=左
	ComboCount   int32   // 连击次数
	Invincible   bool    // 是否无敌

	// 攻击帧跟踪
	AttackStarted bool            // 是否已开始攻击
	AttackHits    map[string]bool // 已命中的玩家ID
}

// startFightStage 启动战斗阶段
func (r *Room) startFightStage() {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_FIGHT
	r.mu.Unlock()
	r.log("战斗阶段开始")

	// 初始化第一回合
	r.startRound()
}

// startRound 开始新回合
func (r *Room) startRound() {
	r.RoundNo++
	r.FrameNo = 0
	r.RoundWinner = ""
	r.RoundTimeout = false
	r.RoundEnded = false
	r.log("===== 回合 %d 开始 =====", r.RoundNo)

	// 初始化本回合实体
	r.mu.Lock()
	entities := r.createRoundEntities()
	r.mu.Unlock()

	// 广播RoundResult
	r.broadcastStageState(game.GameStage_STAGE_FIGHT, r.cfg.FightTimeout)

	// 启动帧循环
	r.fightTick = time.NewTicker(frameInterval)
	defer r.fightTick.Stop()

	timeoutFrames := int32(r.cfg.FightTimeout * fps)
	var entityMu sync.Mutex

	for range r.fightTick.C {
		select {
		case <-r.closeCh:
			return
		default:
		}

		r.FrameNo++

		// 1. 收集输入
		for _, p := range r.Players {
			input, _, ok := p.ConsumeInput()
			if ok {
				entityMu.Lock()
				r.applyInput(entities, p.PlayerId, input)
				entityMu.Unlock()
			}
		}

		// 2. 物理更新
		entityMu.Lock()
		r.updatePhysics(entities)

		// 3. 碰撞检测
		r.checkCollisions(entities)
		entityMu.Unlock()

		// 4. 广播帧状态
		snapshot := r.buildSnapshot(entities)
		data, err := proto.Marshal(snapshot)
		if err == nil {
			r.broadcast(data)
		}

		// 5. 检查回合结束
		if r.RoundEnded {
			r.endRound(entities)
			return
		}

		// 6. 检查超时
		if r.FrameNo >= timeoutFrames {
			r.RoundTimeout = true
			r.RoundEnded = true
			r.endRound(entities)
			return
		}
	}
}

// createRoundEntities 创建本回合的实体
func (r *Room) createRoundEntities() map[string]*Entity {
	entities := make(map[string]*Entity)

	spawnX := []float64{200, 600}

	for i, p := range r.Players {
		hp := p.MaxHP
		if r.RoundNo > 1 && p.PlayerId == r.RoundWinner {
			// 胜者继承剩余血量
			hp = p.CurrentHP
		} else if r.RoundNo > 1 {
			// 败者切换下一个人物
			if p.CurrentCharacterIdx < len(p.RemainingCharacters)-1 {
				p.CurrentCharacterIdx++
			} else {
				// 已无人可用，游戏结束
				r.GameResult = &game.GameResult{
					RoomId:   r.RoomId,
					WinnerId: r.opponentOf(p.PlayerId).PlayerId,
					LoserId:  p.PlayerId,
				}
				return entities
			}
		}

		// 如果第一回合或胜者，使用当前角色
		if r.RoundNo == 1 {
			p.CurrentCharacterIdx = 0
		}
		p.ResetForNewRound(hp)
		p.CurrentHP = hp

		entities[p.PlayerId] = &Entity{
			PlayerId:    p.PlayerId,
			CharacterId: p.CurrentCharacterId(),
			X:           spawnX[i],
			Y:           groundY,
			HP:          hp,
			MaxHP:       p.MaxHP,
			ActionState: ActionIdle,
			Facing:      float64(1 - 2*i), // P1向右, P2向左
			AttackHits:  make(map[string]bool),
		}
	}

	return entities
}

// applyInput 应用玩家输入
func (r *Room) applyInput(entities map[string]*Entity, playerId string, actionMask int32) {
	e, ok := entities[playerId]
	if !ok {
		return
	}

	// 受击/倒地状态不可操作
	if e.ActionState == ActionHit || e.ActionState == ActionDown || e.ActionState == ActionBlockstun {
		return
	}

	// 前摇/后摇中不能打断
	if e.ActionState >= ActionAttackLight && e.ActionState <= ActionSkill {
		if r.FrameNo-e.FrameCounter < 3 { // 简单后摇
			return
		}
	}

	e.ActionState = ActionIdle
	e.Vx = 0

	if actionMask&ActionMaskLeft != 0 {
		e.Vx = -moveSpeed
		e.ActionState = ActionWalk
		e.Facing = -1
	}
	if actionMask&ActionMaskRight != 0 {
		e.Vx = moveSpeed
		e.ActionState = ActionWalk
		e.Facing = 1
	}
	if actionMask&ActionMaskJump != 0 && e.Y >= groundY {
		e.Vy = jumpVelocity
		e.ActionState = ActionJump
	}
	if actionMask&ActionMaskGuard != 0 {
		e.ActionState = ActionGuard
		e.Vx = 0
	}
	if actionMask&ActionMaskEvade != 0 {
		e.ActionState = ActionEvade
		e.Vx = evadeSpeed * e.Facing
		e.Invincible = true
		r.log("玩家 %s 闪避", playerId)
	}
	if actionMask&ActionMaskLight != 0 {
		e.ActionState = ActionAttackLight
		e.FrameCounter = r.FrameNo
		e.AttackStarted = true
		e.AttackHits = make(map[string]bool)
	}
	if actionMask&ActionMaskHeavy != 0 {
		e.ActionState = ActionAttackHeavy
		e.FrameCounter = r.FrameNo
		e.AttackStarted = true
		e.AttackHits = make(map[string]bool)
	}
	if actionMask&ActionMaskSkill != 0 {
		e.ActionState = ActionSkill
	}
}

// updatePhysics 更新物理
func (r *Room) updatePhysics(entities map[string]*Entity) {
	for _, e := range entities {
		// 重力
		if e.Y < groundY {
			e.Vy += gravity
			e.Y += e.Vy
			if e.Y >= groundY {
				e.Y = groundY
				e.Vy = 0
				if e.ActionState == ActionJump {
					e.ActionState = ActionIdle
				}
			}
		}

		// 水平移动
		e.X += e.Vx

		// 边界
		if e.X < minX {
			e.X = minX
		}
		if e.X > maxX {
			e.X = maxX
		}

		// 闪避结束
		if e.ActionState == ActionEvade {
			if r.FrameNo-e.FrameCounter >= evadeDuration {
				e.Invincible = false
				e.ActionState = ActionIdle
			}
		}

		// 受击恢复
		if e.ActionState == ActionHit {
			if r.FrameNo-e.FrameCounter >= heavyHitstun {
				e.ActionState = ActionIdle
			}
		}
		if e.ActionState == ActionBlockstun {
			if r.FrameNo-e.FrameCounter >= 8 {
				e.ActionState = ActionIdle
			}
		}
	}
}

// checkCollisions 碰撞检测
func (r *Room) checkCollisions(entities map[string]*Entity) {
	for _, atk := range entities {
		// 检查攻击命中
		isAttacking := atk.ActionState == ActionAttackLight || atk.ActionState == ActionAttackHeavy
		if !isAttacking || !atk.AttackStarted {
			continue
		}

		// 攻击前摇
		startup := lightStartup
		dmg := lightDamage
		rng := lightRange
		if atk.ActionState == ActionAttackHeavy {
			startup = heavyStartup
			dmg = heavyDamage
			rng = heavyRange
		}

		if r.FrameNo-atk.FrameCounter < int32(startup) {
			continue
		}

		for _, target := range entities {
			if target.PlayerId == atk.PlayerId {
				continue
			}
			if target.Invincible {
				continue
			}
			if atk.AttackHits[target.PlayerId] {
				continue
			}

			// 简单距离检测
			dist := math.Abs(atk.X - target.X)
			if dist > rng {
				continue
			}
			if math.Abs(atk.Y-target.Y) > 60 {
				continue
			}

			// 命中
			atk.AttackHits[target.PlayerId] = true

			if target.ActionState == ActionGuard {
				// 防御成功
				reducedDmg := int32(float64(dmg) * guardDamageReduce)
				target.HP -= reducedDmg
				target.Vx = guardKnockback * -atk.Facing
				target.ActionState = ActionBlockstun
				target.FrameCounter = r.FrameNo
				r.log("玩家 %s 防御, 受到 %d 点伤害", target.PlayerId, reducedDmg)
			} else {
				// 正常命中
				target.HP -= int32(dmg)
				target.Vx = 6 * -atk.Facing
				target.ActionState = ActionHit
				target.FrameCounter = r.FrameNo
				atk.ComboCount++
				r.log("玩家 %s 命中 %s, 伤害 %d, 剩余HP %d",
					atk.PlayerId, target.PlayerId, dmg, target.HP)
			}

			// 检查KO
			if target.HP <= 0 {
				target.HP = 0
				target.ActionState = ActionDown
				r.RoundWinner = atk.PlayerId
				r.RoundEnded = true
				r.log("KO! 玩家 %s 获胜本回合!", atk.PlayerId)
				return
			}
		}
	}
}

// buildSnapshot 构建帧快照
func (r *Room) buildSnapshot(entities map[string]*Entity) *game.FrameSnapshot {
	states := make([]*game.EntityState, 0, len(entities))
	for _, e := range entities {
		states = append(states, &game.EntityState{
			EntityId:     e.PlayerId,
			X:            float32(e.X),
			Y:            float32(e.Y),
			Vx:           float32(e.Vx),
			Vy:           float32(e.Vy),
			Hp:           e.HP,
			MaxHp:        e.MaxHP,
			ActionState:  e.ActionState,
			FrameCounter: r.FrameNo,
			ComboCount:   e.ComboCount,
			Invincible:   e.Invincible,
			Facing:       float32(e.Facing),
		})
	}
	return &game.FrameSnapshot{
		FrameNo:  r.FrameNo,
		Entities: states,
		RoundNo:  r.RoundNo,
	}
}

// endRound 结束当前回合
func (r *Room) endRound(entities map[string]*Entity) {
	// 超时判定：血量多者胜
	if r.RoundTimeout && r.RoundWinner == "" {
		var maxHP int32 = 0
		for _, e := range entities {
			if e.HP > maxHP {
				maxHP = e.HP
				r.RoundWinner = e.PlayerId
			}
		}
		r.log("回合超时，%s 血量领先获胜", r.RoundWinner)
	}

	// 记录回合结果
	loser := r.opponentOf(r.RoundWinner)
	result := &game.RoundResult{
		RoundNo:        r.RoundNo,
		WinnerId:       r.RoundWinner,
		LoserId:        loser.PlayerId,
		WinnerRemainHp: entities[r.RoundWinner].HP,
		RoundTimeout: func() int32 {
			if r.RoundTimeout {
				return 1
			}
			return 0
		}(),
	}
	r.RoundResults = append(r.RoundResults, result)

	// 更新胜者血量
	for _, p := range r.Players {
		if p.PlayerId == r.RoundWinner {
			p.CurrentHP = entities[r.RoundWinner].HP
		}
	}

	// 广播回合结果
	data, _ := proto.Marshal(result)
	if data != nil {
		r.broadcast(data)
	}
	r.log("回合 %d 结束, 胜者: %s", r.RoundNo, r.RoundWinner)

	// 检查游戏结束
	r.checkGameEnd(entities)
}

// checkGameEnd 检查游戏是否结束
func (r *Room) checkGameEnd(entities map[string]*Entity) {
	// 检查败者是否还有可用角色
	for _, p := range r.Players {
		if p.PlayerId != r.RoundWinner {
			if p.CurrentCharacterIdx >= len(p.RemainingCharacters)-1 {
				// 没有下一个人物了
				r.endGame(r.RoundWinner)
				return
			}
		}
	}

	// 继续下一回合
	r.startRound()
}

// endGame 结束整局游戏
func (r *Room) endGame(winnerId string) {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_SETTLEMENT
	loser := r.opponentOf(winnerId)

	r.GameResult = &game.GameResult{
		RoomId:      r.RoomId,
		WinnerId:    winnerId,
		LoserId:     loser.PlayerId,
		TotalRounds: r.RoundNo,
		Rounds:      r.RoundResults,
		GameType:    r.GameType,
	}
	r.mu.Unlock()

	r.log("===== 游戏结束! 胜者: %s =====", winnerId)

	// 广播游戏结果
	data, _ := proto.Marshal(r.GameResult)
	if data != nil {
		r.broadcast(data)
	}

	// 结算阶段
	r.startSettlement()
}

// startSettlement 结算阶段
func (r *Room) startSettlement() {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_FIGHT
	r.mu.Unlock()

	settlement := &game.SettlementInfo{
		RoomId:   r.RoomId,
		WinnerId: r.GameResult.WinnerId,
		LoserId:  r.GameResult.LoserId,
	}
	data, _ := proto.Marshal(settlement)
	if data != nil {
		r.broadcast(data)
	}
}
