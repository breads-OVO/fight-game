package strategy

import (
	"context"
	"fight-game/pkg/common/utils"
	"fight-game/service/match/internal/scanner/match"
	"time"
)

// EntertainmentStrategy 娱乐匹配策略 - 简单FIFO，无段位限制
type EntertainmentStrategy struct{}

func (s *EntertainmentStrategy) Match(ctx context.Context, queue match.MatchQueue, repo match.TicketRepository) ([]*match.MatchResult, error) {
	// 1. 尝试弹出2个票
	ticketIDs, err := queue.Pop(ctx, 2)
	if err != nil {
		return nil, err
	}
	if len(ticketIDs) < 2 {
		return nil, nil
	}

	// 2. 获取票信息（用于提取playerID）
	t1, err := repo.GetTicket(ctx, ticketIDs[0])
	if err != nil {
		return nil, err
	}
	t2, err := repo.GetTicket(ctx, ticketIDs[1])
	if err != nil {
		return nil, err
	}

	// 3. 生成房间ID
	roomID := utils.GenUUIDWithPrefix("room_")

	result := &match.MatchResult{
		RoomID:    roomID,
		TicketIDs: ticketIDs,
		PlayerIDs: []string{t1.PlayerId, t2.PlayerId},
		MatchedAt: time.Now().Unix(),
	}

	// 4. 原子更新状态（Lua脚本内部已完成：更新票状态 + 删除玩家匹配键）
	if err := repo.UpdateMatched(ctx, result); err != nil {
		return nil, err
	}

	return []*match.MatchResult{result}, nil
}
