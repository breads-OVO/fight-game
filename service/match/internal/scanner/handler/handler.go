package handler

import (
	"context"
	"fmt"
	"math/rand"

	"fight-game/pb/game"
	"fight-game/service/match/internal/scanner/match"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

// MatchResultHandler 匹配结果处理器，创建房间并通知 Game 服务
type MatchResultHandler struct {
	gameClient game.GameServiceClient
}

func NewMatchResultHandler(client zrpc.Client) *MatchResultHandler {
	return &MatchResultHandler{
		gameClient: game.NewGameServiceClient(client.Conn()),
	}
}

func (h *MatchResultHandler) Handle(ctx context.Context, result *match.MatchResult) error {
	// 生成房间ID
	roomId := fmt.Sprintf("room_%s_%d", result.GameType, rand.Int63())

	// 调用 Game 服务创建房间
	resp, err := h.gameClient.CreateRoom(ctx, &game.CreateRoomRequest{
		RoomId:    roomId,
		GameType:  result.GameType,
		PlayerIds: result.PlayerIDs,
		Rating:    result.Rating,
	})
	if err != nil {
		logx.Errorf("CreateRoom failed: %v, players=%v", err, result.PlayerIDs)
		return fmt.Errorf("create room failed: %w", err)
	}

	// 设置GameAddr供后续写入ticket
	result.RoomID = roomId
	result.GameAddr = resp.WsAddr

	logx.Infof("匹配成功: room=%s players=%v gameAddr=%s", roomId, result.PlayerIDs, resp.WsAddr)
	return nil
}
