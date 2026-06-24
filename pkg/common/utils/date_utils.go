package utils

import "time"

// GetTimestamp 获取当前毫秒级时间戳
func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}
