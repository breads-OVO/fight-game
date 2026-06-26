package strategy

import (
	"context"
	pbQueue "fight-game/pb/match/queue"
	"fight-game/pkg/common/utils"
	"fight-game/service/match/internal/scanner/match"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// CompetitiveStrategy 竞技匹配策略 - 基于段位分(ELO)的扩圈匹配
type CompetitiveStrategy struct {
	initialRange int // 初始分差（如100）
	maxRange     int // 最大分差（如500）
	step         int // 每次扩圈步进（如50）
}

func NewCompetitiveStrategy(initialRange, maxRange, step int) *CompetitiveStrategy {
	return &CompetitiveStrategy{
		initialRange: initialRange,
		maxRange:     maxRange,
		step:         step,
	}
}

// Match 执行竞技匹配：从队列中按 rating 依次扩圈查找匹配对
func (s *CompetitiveStrategy) Match(ctx context.Context, queue match.MatchQueue, repo match.TicketRepository) ([]*match.MatchResult, error) {
	// 从最小分差开始，逐步扩大范围
	for currentRange := s.initialRange; currentRange <= s.maxRange; currentRange += s.step {
		results, err := s.tryMatchInRange(ctx, queue, repo, currentRange)
		if err != nil {
			logx.Errorf("competitive match try range %d error: %v", currentRange, err)
			continue
		}
		if len(results) > 0 {
			return results, nil
		}
	}

	// 最大范围再尝试一次
	return s.tryMatchInRange(ctx, queue, repo, s.maxRange)
}

// tryMatchInRange 从队列弹出最低 rating 的 2 张票，检查分差是否在允许范围内
func (s *CompetitiveStrategy) tryMatchInRange(ctx context.Context, queue match.MatchQueue, repo match.TicketRepository, ratingRange int) ([]*match.MatchResult, error) {
	ticketIDs, err := queue.Pop(ctx, 2)
	if err != nil {
		return nil, err
	}
	if len(ticketIDs) < 2 {
		return nil, nil
	}

	// 获取两张票的信息
	t1, err := repo.GetTicket(ctx, ticketIDs[0])
	if err != nil {
		return nil, err
	}
	t2, err := repo.GetTicket(ctx, ticketIDs[1])
	if err != nil {
		return nil, err
	}

	// 检查 rating 差是否在允许范围内
	diff := t1.RankScore - t2.RankScore
	if diff < 0 {
		diff = -diff
	}
	if diff > ratingRange {
		// 分差过大，将票放回队列（带原 rating 分值）
		_ = queue.PushWithScore(ctx, ticketIDs[0], int64(t1.RankScore))
		_ = queue.PushWithScore(ctx, ticketIDs[1], int64(t2.RankScore))
		return nil, nil
	}

	// 匹配成功，生成房间
	roomID := utils.GenUUIDWithPrefix("room_")
	result := &match.MatchResult{
		RoomID:    roomID,
		TicketIDs: ticketIDs,
		PlayerIDs: []string{t1.PlayerId, t2.PlayerId},
		MatchedAt: time.Now().Unix(),
		GameType:  pbQueue.GameType_COMPETITION,
		Rating:    int32(max(t1.RankScore, t2.RankScore)),
	}

	if err := repo.UpdateMatched(ctx, result); err != nil {
		return nil, err
	}

	return []*match.MatchResult{result}, nil
}
