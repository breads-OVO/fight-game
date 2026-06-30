package room

import (
	"math"
	"math/rand"
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

	// 物理常量
	gravity    = 0.5
	groundY    = 400.0
	minX       = 30.0
	maxX       = 770.0
	groundMinY = 390.0 // 地面判定阈值

	// 防御属性
	guardDamageReduce = 0.3
	guardKnockback    = 30.0

	// 闪避
	evadeDuration = 12
	evadeSpeed    = 8.0

	// 气槽
	meterGainOnHit    int32 = 20  // 命中获得气槽
	meterGainOnHitDef int32 = 10  // 被命中获得气槽
	meterGainOnBlock  int32 = 8   // 防御获得气槽
	meterCostSkill    int32 = 100 // 技能消耗气槽
	meterGainOnAttack int32 = 5   // 攻击挥空获得气槽

	// 连击
	comboTimeout = 20 // 连击超时帧数（超过此帧数未命中则连击中断）

	// 倒地
	downRecoveryFrames = 30 // 倒地恢复帧数

	// 击退衰减
	knockbackDecay = 0.85 // 每帧击退速度衰减系数
)

// Projectile 飞行道具
type Projectile struct {
	EntityId   string
	X, Y       float64
	Vx, Vy     float64
	OwnerId    string
	Damage     int32
	Range      float64
	Active     bool
	LifeFrames int32 // 剩余存活帧
	HitTargets map[string]bool
}

// Entity 战场实体
type Entity struct {
	PlayerId     string
	CharacterId  string
	X, Y         float64
	Vx, Vy       float64
	HP           int32
	MaxHP        int32
	ActionState  int32
	FrameCounter int32
	Facing       float64 // 1=右, -1=左
	Invincible   bool
	ComboCount   int32

	// 攻击帧跟踪
	AttackStarted bool
	AttackHits    map[string]bool

	// 引用玩家数据（方便访问气槽、角色属性等）
	PlayerData *PlayerData
}

// fightLoop 战斗循环（在独立 goroutine 中运行）
func (r *Room) fightLoop() {
	defer func() {
		if r.onCleanup != nil {
			r.onCleanup()
		}
	}()

	// 第一回合
	r.startRound()

	// 逐回合进行，直到游戏结束
	for {
		// 等待下一回合信号或游戏结束
		select {
		case _, ok := <-r.nextRoundCh:
			if !ok {
				// channel 关闭，游戏结束
				return
			}
			r.startRound()
		case <-r.closeCh:
			return
		}
	}
}

// startFightStage 启动战斗阶段
func (r *Room) startFightStage() {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_FIGHT
	r.mu.Unlock()
	r.log("战斗阶段开始")

	go r.fightLoop()
}

// startRound 开始新回合（阻塞直到回合结束）
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

	// 检查是否因无人可用而导致entities为空
	if entities == nil || len(entities) == 0 {
		return
	}

	// 广播阶段状态
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

		// 重置实体输入状态
		entityMu.Lock()
		for _, e := range entities {
			e.PlayerData.ComboTimer--
			if e.PlayerData.ComboTimer <= 0 {
				e.PlayerData.ComboCount = 0
			}
		}

		// 1. 收集并应用输入
		for _, p := range r.Players {
			input, _, ok := p.ConsumeInput()
			if ok {
				r.applyInput(entities, p.PlayerId, input)
			}
		}

		// 2. 物理更新
		r.updatePhysics(entities)

		// 3. 更新飞行道具
		r.updateProjectiles(entities)

		// 4. 碰撞检测
		r.checkCollisions(entities)
		entityMu.Unlock()

		// 5. 检查断线判负
		if r.checkDisconnectForfeit(entities) {
			r.endRound(entities)
			return
		}

		// 6. 检查回合结束
		if r.RoundEnded {
			r.endRound(entities)
			return
		}

		// 7. 检查超时
		if r.FrameNo >= timeoutFrames {
			r.RoundTimeout = true
			r.RoundEnded = true
			r.endRound(entities)
			return
		}

		// 8. 广播帧状态
		snapshot := r.buildSnapshot(entities)
		data, err := proto.Marshal(snapshot)
		if err == nil {
			r.broadcast(data)
		}
	}
}

// createRoundEntities 创建本回合的实体
func (r *Room) createRoundEntities() map[string]*Entity {
	spawnX := []float64{200, 600}
	entities := make(map[string]*Entity)

	for i, p := range r.Players {
		stats := p.CurrentCharacterStats()
		if stats == nil {
			r.log("玩家 %s 没有可用角色", p.PlayerId)
			continue
		}

		hp := stats.MaxHP
		if r.RoundNo > 1 && p.PlayerId == r.RoundWinner {
			// 胜者继承剩余血量
			hp = p.CurrentHP
		} else if r.RoundNo > 1 {
			// 败者切换下一个人物
			p.ResetMeterForNewCharacter()
			if !p.NextCharacter() {
				// 已无人可用，游戏结束
				r.GameResult = &game.GameResult{
					RoomId:   r.RoomId,
					WinnerId: r.opponentOf(p.PlayerId).PlayerId,
					LoserId:  p.PlayerId,
				}
				return nil
			}
			stats = p.CurrentCharacterStats()
			hp = stats.MaxHP
		}

		p.ResetForNewRound(hp)
		p.CurrentHP = hp

		entities[p.PlayerId] = &Entity{
			PlayerId:    p.PlayerId,
			CharacterId: p.CurrentCharacterId(),
			X:           spawnX[i],
			Y:           groundY,
			HP:          hp,
			MaxHP:       stats.MaxHP,
			ActionState: ActionIdle,
			Facing:      float64(1 - 2*i),
			AttackHits:  make(map[string]bool),
			PlayerData:  p,
		}

		r.log("玩家 %s 出战角色: %s (HP:%d)", p.PlayerId, stats.Name, hp)
	}

	return entities
}

// applyInput 应用玩家输入
func (r *Room) applyInput(entities map[string]*Entity, playerId string, actionMask int32) {
	e, ok := entities[playerId]
	if !ok {
		return
	}

	stats := e.PlayerData.CurrentCharacterStats()
	if stats == nil {
		return
	}

	// 受击/倒地状态不可操作
	if e.PlayerData.HitstunRemaining > 0 || e.PlayerData.BlockstunRemain > 0 || e.PlayerData.IsDown {
		return
	}
	// 受击动作状态限制
	if e.ActionState == ActionHit || e.ActionState == ActionDown {
		return
	}

	// 攻击后摇中 - 检查是否可以取消（连招系统）
	if e.ActionState >= ActionAttackLight && e.ActionState <= ActionSkill {
		framesSinceAttack := r.FrameNo - e.FrameCounter
		canCancel := false

		if e.ActionState == ActionAttackLight && framesSinceAttack >= stats.LightStartup+stats.LightActive {
			// 轻击后摇期间可取消为重击或技能
			if actionMask&ActionMaskHeavy != 0 || actionMask&ActionMaskSkill != 0 {
				canCancel = true
			}
		}
		if e.ActionState == ActionAttackHeavy && framesSinceAttack >= stats.HeavyStartup+stats.HeavyActive {
			// 重击后摇期间可取消为技能
			if actionMask&ActionMaskSkill != 0 {
				canCancel = true
			}
		}
		if e.ActionState == ActionSkill {
			// 技能后摇期间不可取消
			skillDef := skillDB[e.CharacterId]
			if skillDef != nil && framesSinceAttack < skillDef.Startup+skillDef.Active+skillDef.Recovery {
				return
			}
			return // 技能后摇结束后才能行动
		}

		// 简单后摇锁定（不可取消时）
		if !canCancel && framesSinceAttack < 4 {
			return
		}
	}

	// 闪避状态不能操作
	if e.ActionState == ActionEvade {
		return
	}

	// 重置状态
	if e.ActionState != ActionJump {
		e.ActionState = ActionIdle
	}
	e.Vx = 0

	// --- 技能（优先判断） ---
	if actionMask&ActionMaskSkill != 0 && e.PlayerData.CanUseSkill() {
		skillDef := skillDB[e.CharacterId]
		if skillDef != nil {
			e.ActionState = ActionSkill
			e.FrameCounter = r.FrameNo
			e.AttackStarted = true
			e.AttackHits = make(map[string]bool)
			e.PlayerData.Meter -= skillDef.MeterCost * 100
			e.PlayerData.SkillCooldown = 30 // 30帧冷却

			r.log("玩家 %s 释放技能: %s (伤害:%d 范围:%.0f)",
				playerId, skillDef.Name, skillDef.Damage, skillDef.Range)
			return
		}
	}

	// --- 移动 ---
	if actionMask&ActionMaskLeft != 0 {
		e.Vx = -stats.MoveSpeed
		e.ActionState = ActionWalk
		e.Facing = -1
	}
	if actionMask&ActionMaskRight != 0 {
		e.Vx = stats.MoveSpeed
		e.ActionState = ActionWalk
		e.Facing = 1
	}

	// --- 跳跃 ---
	if actionMask&ActionMaskJump != 0 && e.Y >= groundMinY {
		e.Vy = stats.JumpVelocity
		e.ActionState = ActionJump
	}

	// --- 防御 ---
	if actionMask&ActionMaskGuard != 0 {
		e.ActionState = ActionGuard
		e.Vx = 0
	}

	// --- 闪避 ---
	if actionMask&ActionMaskEvade != 0 {
		e.ActionState = ActionEvade
		e.FrameCounter = r.FrameNo
		e.Vx = evadeSpeed * e.Facing
		e.Invincible = true
	}

	// --- 轻击 ---
	if actionMask&ActionMaskLight != 0 && e.ActionState != ActionGuard {
		e.ActionState = ActionAttackLight
		e.FrameCounter = r.FrameNo
		e.AttackStarted = true
		e.AttackHits = make(map[string]bool)
		// 挥空气槽
		e.PlayerData.AddMeter(meterGainOnAttack)
	}

	// --- 重击 ---
	if actionMask&ActionMaskHeavy != 0 && e.ActionState != ActionGuard {
		e.ActionState = ActionAttackHeavy
		e.FrameCounter = r.FrameNo
		e.AttackStarted = true
		e.AttackHits = make(map[string]bool)
		e.PlayerData.AddMeter(meterGainOnAttack * 2)
	}
}

// updatePhysics 更新物理
func (r *Room) updatePhysics(entities map[string]*Entity) {
	for _, e := range entities {
		stats := e.PlayerData.CurrentCharacterStats()

		// 重力
		if !e.PlayerData.IsDown && e.Y < groundY {
			e.Vy += gravity
			e.Y += e.Vy
			if e.Y >= groundY {
				e.Y = groundY
				e.Vy = 0
				if e.ActionState == ActionJump {
					e.ActionState = ActionIdle
				}
				e.PlayerData.IsDown = false
			}
		}

		// 水平移动
		e.X += e.Vx

		// 击退衰减
		if e.ActionState == ActionHit || e.ActionState == ActionBlockstun {
			e.Vx *= knockbackDecay
			if math.Abs(e.Vx) < 0.1 {
				e.Vx = 0
			}
		}

		// 边界
		if e.X < minX {
			e.X = minX
			e.Vx = 0
		}
		if e.X > maxX {
			e.X = maxX
			e.Vx = 0
		}

		// 闪避结束
		if e.ActionState == ActionEvade {
			if r.FrameNo-e.FrameCounter >= evadeDuration {
				e.Invincible = false
				e.ActionState = ActionIdle
				e.Vx = 0
			}
		}

		// 受击恢复
		if e.PlayerData.HitstunRemaining > 0 {
			e.PlayerData.HitstunRemaining--
			if e.PlayerData.HitstunRemaining <= 0 {
				e.ActionState = ActionIdle
				e.Vx = 0
			}
		}

		// 防御硬直恢复
		if e.PlayerData.BlockstunRemain > 0 {
			e.PlayerData.BlockstunRemain--
			if e.PlayerData.BlockstunRemain <= 0 {
				e.ActionState = ActionIdle
				e.Vx = 0
			}
		}

		// 倒地恢复
		if e.PlayerData.IsDown {
			e.ActionState = ActionDown
			e.Vx = 0
			e.Invincible = true
			if r.FrameNo-e.FrameCounter >= downRecoveryFrames {
				e.PlayerData.IsDown = false
				e.PlayerData.HitstunRemaining = 0
				e.ActionState = ActionIdle
				e.Invincible = false
				// 倒地起身重置位置
				e.Y = groundY
				r.log("玩家 %s 起身", e.PlayerId)
			}
		}

		// 技能冷却递减
		if e.PlayerData.SkillCooldown > 0 {
			e.PlayerData.SkillCooldown--
		}

		// 气槽自然增长（每帧微量）
		if e.ActionState != ActionHit && e.ActionState != ActionDown {
			e.PlayerData.AddMeter(1)
		}

		// 攻击结束时（过了有效帧仍未命中）清除攻击状态
		if e.AttackStarted && e.ActionState == ActionAttackLight {
			if r.FrameNo-e.FrameCounter >= stats.LightStartup+stats.LightActive+5 {
				e.AttackStarted = false
			}
		}
		if e.AttackStarted && e.ActionState == ActionAttackHeavy {
			if r.FrameNo-e.FrameCounter >= stats.HeavyStartup+stats.HeavyActive+5 {
				e.AttackStarted = false
			}
		}
		if e.AttackStarted && e.ActionState == ActionSkill {
			skillDef := skillDB[e.CharacterId]
			if skillDef != nil && r.FrameNo-e.FrameCounter >= skillDef.Startup+skillDef.Active+skillDef.Recovery {
				e.AttackStarted = false
				e.ActionState = ActionIdle
			}
		}
	}
}

// updateProjectiles 更新飞行道具
func (r *Room) updateProjectiles(entities map[string]*Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, proj := range r.Projectiles {
		if !proj.Active {
			continue
		}

		proj.X += proj.Vx
		proj.Y += proj.Vy
		proj.LifeFrames--

		// 边界消失
		if proj.X < minX || proj.X > maxX || proj.Y < 0 || proj.Y > groundY+50 {
			proj.Active = false
			continue
		}

		// 生命到期
		if proj.LifeFrames <= 0 {
			proj.Active = false
			continue
		}

		// 飞行道具碰撞检测
		for _, e := range entities {
			if e.PlayerId == proj.OwnerId {
				continue
			}
			if proj.HitTargets[e.PlayerId] {
				continue
			}
			if e.Invincible || e.PlayerData.IsDown {
				continue
			}

			dist := math.Abs(proj.X - e.X)
			if dist > proj.Range {
				continue
			}
			if math.Abs(proj.Y-e.Y) > 60 {
				continue
			}

			// 命中
			proj.HitTargets[e.PlayerId] = true
			proj.Active = false // 飞行道具命中后消失

			if e.ActionState == ActionGuard {
				reducedDmg := int32(float64(proj.Damage) * guardDamageReduce)
				e.HP -= reducedDmg
				e.Vx = guardKnockback * -proj.Vx / math.Abs(proj.Vx)
				e.ActionState = ActionBlockstun
				e.PlayerData.BlockstunRemain = 8
				e.PlayerData.AddMeter(meterGainOnBlock)
				r.log("飞行道具命中 %s (防御), 伤害 %d", e.PlayerId, reducedDmg)
			} else {
				e.HP -= proj.Damage
				e.Vx = 8 * -proj.Vx / math.Abs(proj.Vx)
				e.ActionState = ActionHit
				e.PlayerData.HitstunRemaining = 12
				e.PlayerData.AddMeter(meterGainOnHitDef)
				r.log("飞行道具命中 %s, 伤害 %d", e.PlayerId, proj.Damage)
			}

			// 检查KO
			if e.HP <= 0 {
				e.HP = 0
				e.ActionState = ActionDown
				e.PlayerData.IsDown = true
				e.FrameCounter = r.FrameNo
				r.RoundWinner = proj.OwnerId
				r.RoundEnded = true
				r.log("KO! 玩家 %s 被飞行道具击倒!", e.PlayerId)
				return
			}
		}
	}
}

// checkCollisions 碰撞检测
func (r *Room) checkCollisions(entities map[string]*Entity) {
	for _, atk := range entities {
		stats := atk.PlayerData.CurrentCharacterStats()
		if stats == nil {
			continue
		}

		// 检查攻击命中
		if !atk.AttackStarted {
			continue
		}

		// 获取当前攻击的属性
		var startup, active, dmg int32
		var rng float64
		var hitstun int32
		isSkill := false

		switch atk.ActionState {
		case ActionAttackLight:
			startup = stats.LightStartup
			dmg = stats.LightDamage
			rng = stats.LightRange
			active = stats.LightActive
			hitstun = stats.LightHitstun
		case ActionAttackHeavy:
			startup = stats.HeavyStartup
			dmg = stats.HeavyDamage
			rng = stats.HeavyRange
			active = stats.HeavyActive
			hitstun = stats.HeavyHitstun
		case ActionSkill:
			skillDef := skillDB[atk.CharacterId]
			if skillDef != nil {
				startup = skillDef.Startup
				dmg = skillDef.Damage
				rng = skillDef.Range
				active = skillDef.Active
				hitstun = skillDef.Hitstun
				isSkill = true
			}
		default:
			continue
		}

		// 前摇阶段
		elapsed := r.FrameNo - atk.FrameCounter
		if elapsed < startup {
			continue
		}

		// 有效帧窗口
		hitActiveWindow := elapsed - startup
		if hitActiveWindow >= active {
			// 技能飞行道具
			if isSkill {
				skillDef := skillDB[atk.CharacterId]
				if skillDef != nil && skillDef.Special == 3 {
					// 生成飞行道具
					proj := &Projectile{
						EntityId:   atk.PlayerId + "_proj_" + string(r.FrameNo),
						X:          atk.X + 30*atk.Facing,
						Y:          atk.Y - 20,
						Vx:         skillDef.ProjectileVx * atk.Facing,
						Vy:         0,
						OwnerId:    atk.PlayerId,
						Damage:     skillDef.Damage,
						Range:      30,
						Active:     true,
						LifeFrames: 60,
						HitTargets: make(map[string]bool),
					}
					r.mu.Lock()
					r.Projectiles = append(r.Projectiles, proj)
					r.mu.Unlock()
					r.log("玩家 %s 发射飞行道具: %s", atk.PlayerId, skillDef.Name)
				} else if skillDef != nil && skillDef.Special == 1 {
					// 升龙（无敌）
					atk.Invincible = true
				}
			}
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
			if target.PlayerData.IsDown {
				continue
			}

			// 距离检测
			dist := math.Abs(atk.X - target.X)
			if dist > rng {
				continue
			}
			if math.Abs(atk.Y-target.Y) > 60 {
				continue
			}

			// 命中
			atk.AttackHits[target.PlayerId] = true

			// 更新连击
			atk.PlayerData.ComboCount++
			atk.PlayerData.ComboTimer = comboTimeout
			atk.PlayerData.LastHitFrame = r.FrameNo

			if target.ActionState == ActionGuard {
				// 防御成功
				reducedDmg := int32(float64(dmg) * guardDamageReduce)
				target.HP -= reducedDmg
				knockbackForce := guardKnockback
				if isSkill {
					skillDef := skillDB[atk.CharacterId]
					if skillDef != nil {
						knockbackForce = skillDef.Knockback * 0.5
					}
				}
				target.Vx = knockbackForce * -atk.Facing
				target.ActionState = ActionBlockstun
				target.PlayerData.BlockstunRemain = hitstun / 2
				target.PlayerData.AddMeter(meterGainOnBlock)

				// 防御方获得气槽
				atk.PlayerData.AddMeter(meterGainOnHitDef)

				r.log("玩家 %s 防御, 受到 %d 点伤害 (连击:%d)",
					target.PlayerId, reducedDmg, atk.PlayerData.ComboCount)
			} else {
				// 正常命中
				target.HP -= dmg
				knockbackForce := 6.0
				if isSkill {
					skillDef := skillDB[atk.CharacterId]
					if skillDef != nil {
						knockbackForce = skillDef.Knockback * 0.1
					}
				}
				target.Vx = knockbackForce * -atk.Facing
				target.ActionState = ActionHit
				target.PlayerData.HitstunRemaining = hitstun
				target.PlayerData.AddMeter(meterGainOnHitDef)

				// 攻击方获得气槽
				atk.PlayerData.AddMeter(meterGainOnHit)

				r.log("玩家 %s 命中 %s, 伤害 %d (连击:%d)",
					atk.PlayerId, target.PlayerId, dmg, atk.PlayerData.ComboCount)
			}

			// 检查KO
			if target.HP <= 0 {
				target.HP = 0
				target.ActionState = ActionDown
				target.PlayerData.IsDown = true
				target.PlayerData.HitstunRemaining = 0
				target.FrameCounter = r.FrameNo
				r.RoundWinner = atk.PlayerId
				r.RoundEnded = true
				r.log("KO! 玩家 %s 获胜本回合! (连击:%d)", atk.PlayerId, atk.PlayerData.ComboCount)
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
			ComboCount:   e.PlayerData.ComboCount,
			Invincible:   e.Invincible || e.PlayerData.IsDown,
			Facing:       float32(e.Facing),
		})
	}

	// 添加飞行道具到帧快照（如果不为空可以扩展EntityState）
	// 目前实体列表只包含玩家，飞行道具由客户端根据特效显示

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

	// 确保胜者不为空
	if r.RoundWinner == "" {
		r.RoundWinner = r.Players[0].PlayerId
		r.log("回合结束，默认 %s 获胜", r.RoundWinner)
	}

	loser := r.opponentOf(r.RoundWinner)

	// 记录回合结果
	winnerChar := ""
	loserChar := ""
	if entities[r.RoundWinner] != nil {
		winnerChar = entities[r.RoundWinner].CharacterId
	}
	if entities[loser.PlayerId] != nil {
		loserChar = entities[loser.PlayerId].CharacterId
	}

	result := &game.RoundResult{
		RoundNo:           r.RoundNo,
		WinnerId:          r.RoundWinner,
		LoserId:           loser.PlayerId,
		WinnerCharacterId: winnerChar,
		LoserCharacterId:  loserChar,
		WinnerRemainHp: func() int32 {
			if entities[r.RoundWinner] != nil {
				return entities[r.RoundWinner].HP
			}
			return 0
		}(),
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
		if p.PlayerId == r.RoundWinner && entities[r.RoundWinner] != nil {
			p.CurrentHP = entities[r.RoundWinner].HP
		}
	}

	// 清空飞行道具
	r.mu.Lock()
	r.Projectiles = nil
	r.mu.Unlock()

	// 广播回合结果
	data, _ := proto.Marshal(result)
	if data != nil {
		r.broadcast(data)
	}
	r.log("回合 %d 结束, 胜者: %s", r.RoundNo, r.RoundWinner)

	// 检查游戏是否结束
	if r.checkGameEnd(entities) {
		r.endGame(r.RoundWinner)
	} else {
		// 通知 fightLoop 开始下一回合
		select {
		case r.nextRoundCh <- struct{}{}:
		case <-r.closeCh:
		}
	}
}

// checkGameEnd 检查游戏是否结束
// 返回 true 表示游戏结束
func (r *Room) checkGameEnd(entities map[string]*Entity) bool {
	// 检查败者是否还有可用角色
	for _, p := range r.Players {
		if p.PlayerId != r.RoundWinner {
			if p.CurrentCharacterIdx >= len(p.RemainingCharacters)-1 {
				r.log("玩家 %s 已无可用角色，游戏结束", p.PlayerId)
				return true
			}
		}
	}
	return false
}

// endGame 结束整局游戏
func (r *Room) endGame(winnerId string) {
	r.mu.Lock()
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

	// 进入结算阶段
	r.startSettlement()

	// 关闭 nextRoundCh 停止 fightLoop
	close(r.nextRoundCh)
}

// startSettlement 结算阶段
func (r *Room) startSettlement() {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_SETTLEMENT
	r.mu.Unlock()

	winnerId := r.GameResult.WinnerId
	loserId := r.GameResult.LoserId

	// 计算剩余角色数
	winnerRemainChamps := int32(0)
	for _, p := range r.Players {
		if p.PlayerId == winnerId {
			winnerRemainChamps = int32(len(p.RemainingCharacters) - p.CurrentCharacterIdx)
		}
	}

	settlement := &game.SettlementInfo{
		RoomId:             r.RoomId,
		WinnerId:           winnerId,
		LoserId:            loserId,
		WinnerRemainChamps: winnerRemainChamps,
	}
	data, _ := proto.Marshal(settlement)
	if data != nil {
		r.broadcast(data)
	}

	// 排名分计算
	if r.GameType == 1 { // 竞技模式
		winnerRatingDelta := int32(20 + rand.Int31n(10)) // 20~29
		loserRatingDelta := int32(-15 - rand.Int31n(10)) // -15~-24

		r.log("竞技结算: %s 段位分 %+d, %s 段位分 %+d",
			winnerId, winnerRatingDelta, loserId, loserRatingDelta)

		// 异步通知 Player 服务更新段位分
		// TODO: 通过 gRPC 调用 PlayerService.UpdateRating/UpdateMatchStats
		_ = winnerRatingDelta
		_ = loserRatingDelta
	}

	r.log("结算完成, 胜者: %s, 剩余角色: %d", winnerId, winnerRemainChamps)

	// 触发房间清理（延迟一点，确保消息发送完成）
	time.AfterFunc(100*time.Millisecond, func() {
		if r.onCleanup != nil {
			r.log("房间 %s 清理中...", r.RoomId)
			r.onCleanup()
		}
	})
}

// checkDisconnectForfeit 检查是否有玩家断线超时，超时则对手获胜
func (r *Room) checkDisconnectForfeit(entities map[string]*Entity) bool {
	for _, p := range r.Players {
		if p.Connected || p.DisconnectTime.IsZero() {
			continue
		}
		if time.Since(p.DisconnectTime) >= disconnectTimeout {
			winner := r.opponentOf(p.PlayerId)
			r.log("玩家 %s 断线超过 %v，%s 获胜", p.PlayerId, disconnectTimeout, winner.PlayerId)
			r.RoundWinner = winner.PlayerId
			r.RoundEnded = true
			return true
		}
	}
	return false
}
