package match

import "context"

// MatchStrategy 定义匹配算法
type MatchStrategy interface {

	// Match 执行匹配，返回一个或多个匹配结果；若无法匹配返回 nil, nil
	Match(ctx context.Context, queue MatchQueue, repo TicketRepository) ([]*MatchResult, error)
}
