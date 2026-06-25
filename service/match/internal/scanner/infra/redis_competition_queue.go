package infra

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisCompetitionQueue 竞技匹配队列，基于 Redis ZSet 实现
type RedisCompetitionQueue struct {
	rdb redis.UniversalClient
	key string
}

func NewRedisCompetitionQueue(rdb redis.UniversalClient, key string) *RedisCompetitionQueue {
	return &RedisCompetitionQueue{rdb: rdb, key: key}
}

// Pop 原子弹出 count 个最低 rating 的票（ZSet 按 score 升序）
// 使用 Lua 脚本保证 ZRANGEBYSCORE + ZREM 原子性
func (q *RedisCompetitionQueue) Pop(ctx context.Context, count int) ([]string, error) {
	script := `
		local members = redis.call('ZRANGEBYSCORE', KEYS[1], 0, 999999, 'LIMIT', 0, ARGV[1])
		if #members == 0 then
			return {}
		end
		for _, member in ipairs(members) do
			redis.call('ZREM', KEYS[1], member)
		end
		return members
	`
	vals, err := q.rdb.Eval(ctx, script, []string{q.key}, count).Result()
	if err != nil {
		return nil, err
	}
	if vals == nil {
		return []string{}, nil
	}
	if slice, ok := vals.([]interface{}); ok {
		res := make([]string, len(slice))
		for i, v := range slice {
			res[i] = v.(string)
		}
		return res, nil
	}
	return []string{}, nil
}

func (q *RedisCompetitionQueue) Len(ctx context.Context) (int64, error) {
	return q.rdb.ZCard(ctx, q.key).Result()
}

func (q *RedisCompetitionQueue) Push(ctx context.Context, id string) error {
	// 竞技队列必须带 score，Push 作为降级默认用 0
	return q.rdb.ZAdd(ctx, q.key, redis.Z{Score: 0, Member: id}).Err()
}

func (q *RedisCompetitionQueue) PushWithScore(ctx context.Context, id string, score int64) error {
	return q.rdb.ZAdd(ctx, q.key, redis.Z{Score: float64(score), Member: id}).Err()
}

func (q *RedisCompetitionQueue) Remove(ctx context.Context, id string) error {
	return q.rdb.ZRem(ctx, q.key, id).Err()
}
