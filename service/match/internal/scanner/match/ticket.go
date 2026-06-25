package match

import (
	"context"
	"fight-game/pb/match/queue"
)

// TicketRepository 管理票的存储
type TicketRepository interface {

	// CreateTicket 创建票并存储到 Redis Hash
	CreateTicket(ctx context.Context, ticket *Ticket) error

	// GetTicket 获取票详情
	GetTicket(ctx context.Context, ticketID string) (*Ticket, error)

	// UpdateTicketStatus 更新票状态（可附带其他字段）
	UpdateTicketStatus(ctx context.Context, ticketID string, status queue.MatchStatus, extra map[string]interface{}) error

	// DeleteTicket 删除票（匹配完成后清理）
	DeleteTicket(ctx context.Context, ticketID string) error

	// SetPlayerTicket 设置玩家->票映射，返回是否成功（false表示玩家已在队列中）
	SetPlayerTicket(ctx context.Context, playerID string, ticketID string) (bool, error)

	// GetPlayerTicket 获取玩家当前排队中的票ID（空字符串表示不在队列中）
	GetPlayerTicket(ctx context.Context, playerID string) (string, error)

	// RemovePlayerMatch 删除玩家匹配状态（防止重复入队）
	RemovePlayerMatch(ctx context.Context, playerID string) error

	// UpdateMatched 原子更新多张票为匹配成功，并删除玩家匹配键（Lua脚本保证原子性）
	UpdateMatched(ctx context.Context, result *MatchResult) error
}
