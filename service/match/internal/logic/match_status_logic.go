package logic

import (
	"context"
	"strings"

	"fight-game/pb/match/queue"
	"fight-game/service/match/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMatchStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MatchStatusLogic {
	return &MatchStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetMatchStatus 获取匹配状态
func (l *MatchStatusLogic) GetMatchStatus(in *queue.MatchStatusRequest) (*queue.MatchStatusResponse, error) {
	// 1. 根据 ticket_id 查询 Hash 中的票详情
	ticket, err := l.svcCtx.Repo.GetTicket(l.ctx, in.TicketId)
	if err != nil {
		logx.Errorf("get ticket %s error: %v", in.TicketId, err)
		return nil, status.Error(codes.NotFound, "ticket not found")
	}

	resp := &queue.MatchStatusResponse{
		Status: ticket.Status,
	}

	// 2. 如果已匹配成功，从 Hash 中读取房间信息
	if ticket.Status == queue.MatchStatus_MATCHED {
		// 从 Redis 额外读取 roomId 和 playerIds（存于同一个 Hash 中）
		// ticket 只包含基础字段，roomId/playerIds 在 extra data 中
		// 直接通过 Redis Hash 读取完整数据
		ticketData, err := l.svcCtx.Repo.GetTicket(l.ctx, in.TicketId)
		if err == nil {
			_ = ticketData // 已有基础信息
		}

		// 从 repository 获取额外信息：使用 raw Redis 读取 roomId, playerIds, gameAddr
		rdb := l.svcCtx.Redis
		vals, err := rdb.HMGet(l.ctx, "match:ticket:"+in.TicketId, "roomId", "playerIds", "gameAddr").Result()
		if err == nil {
			if len(vals) >= 1 && vals[0] != nil {
				resp.RoomId = vals[0].(string)
			}
			if len(vals) >= 2 && vals[1] != nil {
				playerIdsStr := vals[1].(string)
				resp.PlayerIds = strings.Split(playerIdsStr, ",")
			}
			if len(vals) >= 3 && vals[2] != nil {
				resp.GameAddr = vals[2].(string)
			}
		}
	}

	return resp, nil
}
