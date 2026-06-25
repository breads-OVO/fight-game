package infra

import (
	"context"
	"fmt"
	"strconv"

	"fight-game/pb/match/queue"
	"fight-game/service/match/internal/scanner/match"

	"github.com/redis/go-redis/v9"
)

const (
	ticketKeyPrefix = "match:ticket:"
	playerMatchKey  = "match:player:"
)

// RedisTicketRepo 基于 Redis 的票存储实现
type RedisTicketRepo struct {
	rdb redis.UniversalClient
}

func NewRedisTicketRepo(rdb redis.UniversalClient) *RedisTicketRepo {
	return &RedisTicketRepo{rdb: rdb}
}

func (r *RedisTicketRepo) ticketKey(ticketID string) string {
	return ticketKeyPrefix + ticketID
}

// playerKey 玩家匹配键
func (r *RedisTicketRepo) playerKey(playerID string) string {
	return playerMatchKey + playerID
}

// CreateTicket 将票信息存储到 Redis Hash
func (r *RedisTicketRepo) CreateTicket(ctx context.Context, ticket *match.Ticket) error {
	return r.rdb.HSet(ctx, r.ticketKey(ticket.TicketId), map[string]interface{}{
		"playerId":   ticket.PlayerId,
		"status":     int32(ticket.Status),
		"enqueuedAt": ticket.EnqueuedAt,
		"rankScore":  ticket.RankScore,
	}).Err()
}

// SetPlayerTicket 使用 SETNX 设置玩家->票映射，防止重复入队
func (r *RedisTicketRepo) SetPlayerTicket(ctx context.Context, playerID string, ticketID string) (bool, error) {
	return r.rdb.SetNX(ctx, r.playerKey(playerID), ticketID, 0).Result()
}

// GetPlayerTicket 获取玩家当前排队中的票
func (r *RedisTicketRepo) GetPlayerTicket(ctx context.Context, playerID string) (string, error) {
	return r.rdb.Get(ctx, r.playerKey(playerID)).Result()
}

// GetTicket 从 Hash 中获取票详情
func (r *RedisTicketRepo) GetTicket(ctx context.Context, ticketID string) (*match.Ticket, error) {
	data, err := r.rdb.HGetAll(ctx, r.ticketKey(ticketID)).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}

	enqueuedAt, _ := strconv.ParseInt(data["enqueuedAt"], 10, 64)
	rankScore, _ := strconv.Atoi(data["rankScore"])
	statusVal, _ := strconv.Atoi(data["status"])

	return &match.Ticket{
		TicketId:   ticketID,
		PlayerId:   data["playerId"],
		Status:     queue.MatchStatus(statusVal),
		EnqueuedAt: enqueuedAt,
		RankScore:  rankScore,
	}, nil
}

// UpdateTicketStatus 更新票状态
func (r *RedisTicketRepo) UpdateTicketStatus(ctx context.Context, ticketID string, status queue.MatchStatus, extra map[string]interface{}) error {
	fields := map[string]interface{}{
		"status": int32(status),
	}
	for k, v := range extra {
		fields[k] = v
	}
	return r.rdb.HSet(ctx, r.ticketKey(ticketID), fields).Err()
}

// DeleteTicket 删除票
func (r *RedisTicketRepo) DeleteTicket(ctx context.Context, ticketID string) error {
	return r.rdb.Del(ctx, r.ticketKey(ticketID)).Err()
}

// RemovePlayerMatch 删除玩家匹配状态键
func (r *RedisTicketRepo) RemovePlayerMatch(ctx context.Context, playerID string) error {
	return r.rdb.Del(ctx, r.playerKey(playerID)).Err()
}

// UpdateMatched 原子更新匹配结果
// Lua 脚本: 更新票状态为 MATCHED + 房间信息，删除玩家匹配键
func (r *RedisTicketRepo) UpdateMatched(ctx context.Context, result *match.MatchResult) error {
	script := `
		local roomId = ARGV[1]
		local playerIds = ARGV[2]  -- "pid1,pid2"
		for i, ticketKey in ipairs(KEYS) do
			redis.call('HSET', ticketKey, 'status', 1, 'roomId', roomId, 'playerIds', playerIds)
		end
		for i = 3, #ARGV do
			redis.call('DEL', 'match:player:' .. ARGV[i])
		end
		return 1
	`
	keys := make([]string, len(result.TicketIDs))
	for i, tid := range result.TicketIDs {
		keys[i] = r.ticketKey(tid)
	}
	playerIDStr := ""
	for i, pid := range result.PlayerIDs {
		if i > 0 {
			playerIDStr += ","
		}
		playerIDStr += pid
	}
	args := append([]string{result.RoomID, playerIDStr}, result.PlayerIDs...)
	_, err := r.rdb.Eval(ctx, script, keys, args).Result()
	return err
}
