package server

import (
	"context"

	"fight-game/pb/match"
	"fight-game/pb/match/queue"
	"fight-game/service/match/internal/logic"
	"fight-game/service/match/internal/svc"
)

type MatchServiceServer struct {
	svcCtx *svc.ServiceContext
	match.UnimplementedMatchServiceServer
}

func NewMatchServiceServer(svcCtx *svc.ServiceContext) *MatchServiceServer {
	return &MatchServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *MatchServiceServer) JoinQueue(ctx context.Context, in *queue.MatchRequest) (*queue.MatchResponse, error) {
	l := logic.NewJoinQueueLogic(ctx, s.svcCtx)
	return l.JoinQueue(in)
}

func (s *MatchServiceServer) LeaveQueue(ctx context.Context, in *queue.LeaveQueueRequest) (*queue.LeaveQueueResponse, error) {
	l := logic.NewLeaveQueueLogic(ctx, s.svcCtx)
	return l.LeaveQueue(in)
}

func (s *MatchServiceServer) GetMatchStatus(ctx context.Context, in *queue.MatchStatusRequest) (*queue.MatchStatusResponse, error) {
	l := logic.NewMatchStatusLogic(ctx, s.svcCtx)
	return l.GetMatchStatus(in)
}
