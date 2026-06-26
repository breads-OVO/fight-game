package room

import (
	"fight-game/pb/game"
	"time"
)

// startBanStage 启动禁用阶段（仅竞技模式）
// 双方同时各自禁用 cfg.BanCount 个角色
func (r *Room) startBanStage() {
	targetCount := r.banCount
	r.mu.Lock()
	r.stage = game.GameStage_STAGE_BAN
	r.banConfirmed = make(map[string]bool)
	r.mu.Unlock()
	r.log("禁用阶段开始，每人需禁用 %d 个角色", targetCount)

	// 广播阶段状态
	r.broadcastStageState(game.GameStage_STAGE_BAN, r.cfg.BanTimeout)

	// 双方同时禁用，等待完成或超时
	done := make(chan struct{}, 1)
	go func() {
		r.waitBothBanDone()
		done <- struct{}{}
	}()

	timer := time.NewTimer(time.Duration(r.cfg.BanTimeout) * time.Second)

	select {
	case <-done:
		r.log("双方禁用完成")
	case <-timer.C:
		r.log("禁用超时，已ban的填充，未ban的空值")
	case <-r.closeCh:
		timer.Stop()
		return
	}
	timer.Stop()
	// 禁用处理
	r.banHandle()
	// 收集所有被禁角色
	banned := r.collectBannedChars()
	r.log("禁用完成，被禁角色: %v", banned)
	// 进入Pick阶段
	r.startPickStage()
}

// HandleBan 处理玩家禁用请求
func (r *Room) HandleBan(playerId string, characterIds []string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stage != game.GameStage_STAGE_BAN {
		return false
	}

	player := r.GetPlayer(playerId)
	if player == nil {
		return false
	}

	// 保存禁用角色
	player.BannedConfirmed = true
	player.BannedCharacters = characterIds
	r.log("玩家 %s 禁用角色 %d", playerId, characterIds)
	return true
}

// waitBothBanDone 等待双方都完成禁用
func (r *Room) waitBothBanDone() {
	for {
		r.mu.RLock()
		allDone := true
		for _, p := range r.Players {
			if !p.BannedConfirmed {
				allDone = false
				break
			}
		}
		r.mu.RUnlock()

		if allDone {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// banHandle ban位处理
func (r *Room) banHandle() {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 收集已被禁用的角色
	r.banCharacter = r.collectBannedChars()
}

// collectBannedChars 收集所有被禁角色（Pick阶段过滤用）
func (r *Room) collectBannedChars() []string {
	banned := make([]string, 0)
	for _, p := range r.Players {
		banned = append(banned, p.BannedCharacters...)
	}
	return banned
}
