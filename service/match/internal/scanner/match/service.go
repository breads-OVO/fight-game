package match

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// MatchService 匹配服务，协调策略和存储
type MatchService struct {
	queue    MatchQueue
	repo     TicketRepository
	strategy MatchStrategy
	// 可选的匹配结果处理器（如通知玩家）
	handler ResultHandler
}

type ResultHandler interface {
	Handle(ctx context.Context, result *MatchResult) error
}

func NewMatchService(queue MatchQueue, repo TicketRepository, strategy MatchStrategy, handler ResultHandler) *MatchService {
	return &MatchService{
		queue:    queue,
		repo:     repo,
		strategy: strategy,
		handler:  handler,
	}
}

// DoMatch 执行一次匹配尝试
func (s *MatchService) DoMatch(ctx context.Context) {
	results, err := s.strategy.Match(ctx, s.queue, s.repo)
	if err != nil {
		logx.Errorf("match error: %v", err)
		return
	}
	if len(results) == 0 {
		return
	}
	for _, result := range results {
		if err := s.handler.Handle(ctx, result); err != nil {
			logx.Errorf("handle match result error: %v", err)
		}
	}
}
