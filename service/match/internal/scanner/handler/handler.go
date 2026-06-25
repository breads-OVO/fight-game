package handler

import (
	"context"
	"fight-game/service/match/internal/scanner/match"
)

// NoopResultHandler 空匹配结果处理器（仅打印日志，后续可对接Game服务创建房间）
type NoopResultHandler struct{}

func (h *NoopResultHandler) Handle(ctx context.Context, result *match.MatchResult) error {
	// TODO: 调用 Game 服务的 CreateRoom RPC
	// 例如: logx.Infof("match success: room=%s players=%v", result.RoomID, result.PlayerIDs)
	return nil
}
