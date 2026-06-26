package match

import (
	"fight-game/pb/match/queue"
)

// Ticket 匹配票
type Ticket struct {
	TicketId   string            // 匹配票ID
	PlayerId   string            // 玩家ID
	Status     queue.MatchStatus // 匹配状态
	EnqueuedAt int64             // 入队时间
	RankScore  int               // 段位分，0表示无
}

// MatchResult 一次匹配产生的对局结果
type MatchResult struct {
	RoomID    string         // 房间ID
	TicketIDs []string       // 匹配票ID
	PlayerIDs []string       // 玩家ID
	MatchedAt int64          // 匹配时间
	GameType  queue.GameType // 游戏类型
	GameAddr  string         // Game 服务直连地址
	Rating    int32          // 段位分
}
