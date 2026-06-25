package logic

import (
	"context"

	"fight-game/pb/match/queue"
	"fight-game/service/match/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeaveQueueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveQueueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveQueueLogic {
	return &LeaveQueueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LeaveQueue 离开队列
func (l *LeaveQueueLogic) LeaveQueue(in *queue.LeaveQueueRequest) (*queue.LeaveQueueResponse, error) {
	// 1. 获取票信息
	ticket, err := l.svcCtx.Repo.GetTicket(l.ctx, in.TicketId)
	if err != nil {
		logx.Errorf("get ticket %s error: %v", in.TicketId, err)
		return nil, status.Error(codes.NotFound, "ticket not found")
	}

	// 2. 验证票属于该玩家
	if ticket.PlayerId != in.PlayerId {
		return nil, status.Error(codes.PermissionDenied, "ticket does not belong to player")
	}

	// 3. 如果已匹配成功，不允许取消
	if ticket.Status == queue.MatchStatus_MATCHED {
		return nil, status.Error(codes.FailedPrecondition, "already matched, cannot leave")
	}

	// 4. 从对应队列中移除（两个都调用，只有正确的队列会实际移除）
	_ = l.svcCtx.EntertainmentQueue.Remove(l.ctx, in.TicketId)
	_ = l.svcCtx.CompetitionQueue.Remove(l.ctx, in.TicketId)

	// 5. 删除票数据和玩家映射
	_ = l.svcCtx.Repo.DeleteTicket(l.ctx, in.TicketId)
	_ = l.svcCtx.Repo.RemovePlayerMatch(l.ctx, in.PlayerId)

	logx.Infof("player %s left queue, ticket=%s", in.PlayerId, in.TicketId)

	return &queue.LeaveQueueResponse{
		Success: true,
	}, nil
}
