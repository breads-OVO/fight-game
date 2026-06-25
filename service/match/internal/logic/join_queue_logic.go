package logic

import (
	"context"
	"errors"
	"time"

	"fight-game/pb/match/queue"
	"fight-game/pkg/common/utils"
	"fight-game/service/match/internal/scanner/match"
	"fight-game/service/match/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JoinQueueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinQueueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinQueueLogic {
	return &JoinQueueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// JoinQueue 加入匹配队列
func (l *JoinQueueLogic) JoinQueue(in *queue.MatchRequest) (*queue.MatchResponse, error) {
	// 1. 检查玩家是否已在队列中
	existingTicketID, err := l.svcCtx.Repo.GetPlayerTicket(l.ctx, in.PlayerId)
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.Errorf("check player ticket error: %v", err)
		return nil, status.Error(codes.Internal, "check queue failed")
	}
	if existingTicketID != "" {
		return nil, status.Error(codes.AlreadyExists, "already in queue")
	}

	// 2. 生成票ID
	ticketID := utils.GenUUIDWithPrefix("ticket_")
	now := time.Now().Unix()

	// 3. 创建票信息
	rankScore := 0
	var matchType string
	gameType := in.GameType
	if gameType == queue.GameType_COMPETITION {
		// 竞技匹配
		params := in.GetCompetitionMatchParams()
		if params != nil {
			rankScore = int(params.Rating)
		}
	}

	ticket := &match.Ticket{
		TicketId:   ticketID,
		PlayerId:   in.PlayerId,
		Status:     queue.MatchStatus_QUEUEING,
		EnqueuedAt: now,
		RankScore:  rankScore,
	}

	// 4. 存储票到 Redis Hash
	if err := l.svcCtx.Repo.CreateTicket(l.ctx, ticket); err != nil {
		logx.Errorf("create ticket error: %v", err)
		return nil, status.Error(codes.Internal, "create ticket failed")
	}

	// 5. 设置玩家->票映射
	ok, err := l.svcCtx.Repo.SetPlayerTicket(l.ctx, in.PlayerId, ticketID)
	if err != nil {
		logx.Errorf("set player ticket error: %v", err)
		// 回滚：删除已创建的票
		_ = l.svcCtx.Repo.DeleteTicket(l.ctx, ticketID)
		return nil, status.Error(codes.Internal, "set player ticket failed")
	}
	if !ok {
		// 并发冲突，玩家已被其他请求入队
		_ = l.svcCtx.Repo.DeleteTicket(l.ctx, ticketID)
		return nil, status.Error(codes.AlreadyExists, "already in queue")
	}

	// 6. 加入对应队列
	switch gameType {
	case queue.GameType_COMPETITION:
		//竞技匹配
		if err := l.svcCtx.CompetitionQueue.PushWithScore(l.ctx, ticketID, int64(rankScore)); err != nil {
			logx.Errorf("enqueue competition error: %v", err)
			// 回滚
			_ = l.svcCtx.Repo.DeleteTicket(l.ctx, ticketID)
			_ = l.svcCtx.Repo.RemovePlayerMatch(l.ctx, in.PlayerId)
			return nil, status.Error(codes.Internal, "enqueue failed")
		}
	case queue.GameType_ENTERTAINMENT:
		if err := l.svcCtx.EntertainmentQueue.Push(l.ctx, ticketID); err != nil {
			logx.Errorf("enqueue entertainment error: %v", err)
			// 回滚
			_ = l.svcCtx.Repo.DeleteTicket(l.ctx, ticketID)
			_ = l.svcCtx.Repo.RemovePlayerMatch(l.ctx, in.PlayerId)
			return nil, status.Error(codes.Internal, "enqueue failed")
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid game type")
	}

	logx.Infof("player %s joined %s queue, ticket=%s", in.PlayerId, matchType, ticketID)

	return &queue.MatchResponse{
		TicketId:      ticketID,
		EstimatedWait: int32(l.svcCtx.Config.Match.QueueTimeout),
	}, nil
}
