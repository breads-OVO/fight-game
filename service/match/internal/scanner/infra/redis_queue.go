package infra

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisMatchQueue struct {
	rdb redis.UniversalClient
	key string
}

func NewRedisMatchQueue(rdb redis.UniversalClient, key string) *RedisMatchQueue {
	return &RedisMatchQueue{rdb: rdb, key: key}
}

func (q *RedisMatchQueue) Pop(ctx context.Context, count int) ([]string, error) {
	// Lua脚本原子弹出多个元素
	script := `
		local count = tonumber(ARGV[1])
		local results = {}
		for i=1,count do
			local v = redis.call('LPOP', KEYS[1])
			if not v then break end
			table.insert(results, v)
		end
		return results
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

func (q *RedisMatchQueue) Len(ctx context.Context) (int64, error) {
	return q.rdb.LLen(ctx, q.key).Result()
}

func (q *RedisMatchQueue) Push(ctx context.Context, id string) error {
	return q.rdb.LPush(ctx, q.key, id).Err()
}

func (q *RedisMatchQueue) PushWithScore(ctx context.Context, id string, score int64) error {
	// 娱乐队列忽略 score，等同 Push
	return q.rdb.LPush(ctx, q.key, id).Err()
}

func (q *RedisMatchQueue) Remove(ctx context.Context, id string) error {
	return q.rdb.LRem(ctx, q.key, 0, id).Err()
}
