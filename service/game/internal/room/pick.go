package room

import (
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

	//TODO 获取玩家资源角色
	var availableChars []string

	// 如果竞技模式，需要过滤被禁角色
	if r.GameType == pbQueue.GameType_COMPETITION {
		banned := r.banCharacter
		for _, c := range CharacterPool {
			ban := false
			for _, b := range banned {
				if c == b {
					ban = true
					break
				}
			}
			if !ban {
				availableChars = append(availableChars, c)
			}
		}
	} else {
		availableChars = append([]string{}, CharacterPool...)
	}

	r.log("可用角色: %v", availableChars)

	//TODO 向玩家同步可选角色及资源

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
		r.autoPick()
	case <-r.closeCh:
		return
	}

	// 进入战斗阶段
	r.startFightStage()
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

	r.log("玩家 %s 选人完成: chars=%v ", playerId, characterIds)
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
func (r *Room) autoPick() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.Players {
		if p.PickedConfirmed {
			continue
		}

		// TODO 随机选择角色

		r.log("玩家 %s 自动选人: chars=%v", p.PlayerId)
	}
}
