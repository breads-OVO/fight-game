package match

import "context"

// MatchQueue 匹配队列接口
// 娱乐匹配基于 Redis List，竞技匹配基于 Redis ZSet
type MatchQueue interface {

	// Pop 弹出 count 个元素，返回实际弹出的元素列表（可能少于 count）
	Pop(ctx context.Context, count int) ([]string, error)

	// Len 返回队列长度
	Len(ctx context.Context) (int64, error)

	// Push 入队
	Push(ctx context.Context, id string) error

	// PushWithScore 带分值入队（娱乐匹配忽略分值，竞技匹配用作 ZAdd score）
	PushWithScore(ctx context.Context, id string, score int64) error

	// Remove 从队列中移除指定元素
	Remove(ctx context.Context, id string) error
}
