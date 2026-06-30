package room

import (
	"math/rand"

	"fight-game/pb/game"
	pbQueue "fight-game/pb/match/queue"
	"time"
)

// CharacterPool 角色池（完整列表）
var CharacterPool = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

// startPickStage 启动选人阶段
func (r *Room) startPickStage() {
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_PICK
	r.mu.Unlock()

	r.log("选人阶段开始")

	// 广播阶段状态
	r.broadcastStageState(game.GameStage_STAGE_PICK, r.cfg.PickTimeout)

	// 启动倒计时
	timer := time.NewTimer(time.Duration(r.cfg.PickTimeout) * time.Second)
	defer timer.Stop()

	// 筛选可用角色
	availableChars := r.filterAvailableCharacters()

	r.log("可用角色: %d 个", len(availableChars))
	for _, c := range availableChars {
		stats := GetCharacterStats(c)
		r.log("  - %s (%s)", c, stats.Name)
	}

	// 等待选人完成或超时
	done := make(chan struct{}, 1)
	go func() {
		r.waitPickDone()
		done <- struct{}{}
	}()

	select {
	case <-done:
		r.log("选人完成")
	case <-timer.C:
		r.log("选人超时，使用默认选择")
		r.autoPick(availableChars)
	case <-r.closeCh:
		return
	}

	// 进入战斗阶段
	r.startFightStage()
}

// filterAvailableCharacters 筛选可用角色（竞技模式过滤被禁角色）
func (r *Room) filterAvailableCharacters() []string {
	var available []string

	if r.GameType == pbQueue.GameType_COMPETITION {
		banned := r.banCharacter
		for _, c := range CharacterPool {
			isBanned := false
			for _, b := range banned {
				if c == b {
					isBanned = true
					break
				}
			}
			if !isBanned {
				available = append(available, c)
			}
		}
	} else {
		available = append([]string{}, CharacterPool...)
	}

	return available
}

// HandlePick 处理玩家选人请求
func (r *Room) HandlePick(playerId string, pick *game.CharacterConfig) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stage != game.GameStage_STAGE_PICK {
		return false
	}

	player := r.GetPlayer(playerId)
	if player == nil || player.PickedConfirmed {
		return false
	}

	characterIds := pick.CharacterIds
	if len(characterIds) != r.cfg.Characters {
		return false
	}

	player.PickedCharacters = pick
	player.PickedConfirmed = true

	// 初始化战斗数据
	player.RemainingCharacters = make([]string, r.cfg.Characters)
	for i, c := range characterIds {
		player.RemainingCharacters[i] = c
	}

	charNames := make([]string, len(characterIds))
	for i, c := range characterIds {
		stats := GetCharacterStats(c)
		charNames[i] = stats.Name
	}

	r.log("玩家 %s 选人完成: %v", playerId, charNames)
	return true
}

// waitPickDone 等待所有玩家选人完成
func (r *Room) waitPickDone() {
	for {
		allDone := true
		for _, p := range r.Players {
			if !p.PickedConfirmed {
				allDone = false
				break
			}
		}
		if allDone {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// autoPick 超时自动选择
func (r *Room) autoPick(availableChars []string) {
	if len(availableChars) == 0 {
		r.log("无可选角色，使用默认角色池")
		availableChars = CharacterPool
	}

	needChars := r.cfg.Characters
	if needChars <= 0 {
		needChars = 3
	}

	// 确保可用角色足够
	if len(availableChars) < needChars {
		// 不足则补充默认角色
		for _, c := range CharacterPool {
			if len(availableChars) >= needChars {
				break
			}
			has := false
			for _, a := range availableChars {
				if a == c {
					has = true
					break
				}
			}
			if !has {
				availableChars = append(availableChars, c)
			}
		}
	}

	for _, p := range r.Players {
		if p.PickedConfirmed {
			continue
		}

		// 随机打乱并从可用角色中选取
		shuffled := make([]string, len(availableChars))
		copy(shuffled, availableChars)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		selected := shuffled[:needChars]

		p.PickedCharacters = &game.CharacterConfig{
			CharacterIds: selected,
		}
		p.PickedConfirmed = true
		p.RemainingCharacters = make([]string, needChars)
		copy(p.RemainingCharacters, selected)

		charNames := make([]string, len(selected))
		for i, c := range selected {
			charNames[i] = GetCharacterStats(c).Name
		}
		r.log("玩家 %s 自动选人: %v", p.PlayerId, charNames)
	}
}
