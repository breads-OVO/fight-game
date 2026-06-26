package server

import (
	"context"
	"fmt"

	"fight-game/pb/game"
	"fight-game/service/game/internal/svc"
)

type GameServiceServer struct {
	svcCtx *svc.ServiceContext
	game.UnimplementedGameServiceServer
}

func NewGameServiceServer(svcCtx *svc.ServiceContext) *GameServiceServer {
	return &GameServiceServer{
		svcCtx: svcCtx,
	}
}

// CreateRoom 匹配成功后由 Match 服务调用
// CreateRoom 创建游戏房间的方法
// 参数:
//
//	ctx - 上下文信息，用于传递请求的元数据和控制请求的生命周期
//	req - 创建房间的请求，包含房间ID、游戏类型、玩家ID列表和评分等信息
//
// 返回值:
//
//	*game.CreateRoomResponse - 创建房间的响应，包含成功状态、消息和WebSocket地址
//	error - 错误信息，如果创建失败则返回相应错误
func (s *GameServiceServer) CreateRoom(ctx context.Context, req *game.CreateRoomRequest) (*game.CreateRoomResponse, error) {
	// 创建房间并启动
	// 调用服务上下文中的CreateAndStartRoom方法，传入房间ID、游戏类型、玩家ID列表和评分等参数
	room := s.svcCtx.CreateAndStartRoom(req.RoomId, req.GameType, req.PlayerIds, req.GetRating())

	// 启动房间（异步，启动后阶段循环自行运转）
	// 使用goroutine异步启动房间，使房间能够独立运行
	go room.Start()

	// 构建WebSocket地址，用于客户端连接到房间
	// 使用服务配置中的WebSocket地址和房间ID构建完整的连接地址
	wsAddr := fmt.Sprintf("ws://%s/play?roomId=%s", s.svcCtx.Config.Game.WsAddr, req.RoomId)

	// 返回创建成功的响应
	// 包含成功状态、成功消息和WebSocket地址
	return &game.CreateRoomResponse{
		Success: true,
		Message: "room created",
		WsAddr:  wsAddr,
	}, nil
}
