package server

import (
	"context"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/logic"
	"fight-game/service/friend/internal/svc"
)

type FriendServiceServer struct {
	svcCtx *svc.ServiceContext
	friend.UnimplementedFriendServiceServer
}

func NewFriendServiceServer(svcCtx *svc.ServiceContext) *FriendServiceServer {
	return &FriendServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *FriendServiceServer) AddFriend(ctx context.Context, in *friend.AddFriendRequest) (*friend.AddFriendResponse, error) {
	l := logic.NewAddFriendLogic(ctx, s.svcCtx)
	return l.AddFriend(in)
}

func (s *FriendServiceServer) ReplyFriend(ctx context.Context, in *friend.ReplyFriendRequest) (*friend.ReplyFriendResponse, error) {
	l := logic.NewReplyFriendLogic(ctx, s.svcCtx)
	return l.ReplyFriend(in)
}

func (s *FriendServiceServer) RemoveFriend(ctx context.Context, in *friend.RemoveFriendRequest) (*friend.RemoveFriendResponse, error) {
	l := logic.NewRemoveFriendLogic(ctx, s.svcCtx)
	return l.RemoveFriend(in)
}

func (s *FriendServiceServer) GetFriendList(ctx context.Context, in *friend.GetFriendListRequest) (*friend.GetFriendListResponse, error) {
	l := logic.NewGetFriendListLogic(ctx, s.svcCtx)
	return l.GetFriendList(in)
}

func (s *FriendServiceServer) SearchPlayer(ctx context.Context, in *friend.SearchPlayerRequest) (*friend.SearchPlayerResponse, error) {
	l := logic.NewSearchPlayerLogic(ctx, s.svcCtx)
	return l.SearchPlayer(in)
}

func (s *FriendServiceServer) SendChatMessage(ctx context.Context, in *friend.SendChatMessageRequest) (*friend.SendChatMessageResponse, error) {
	l := logic.NewSendChatMessageLogic(ctx, s.svcCtx)
	return l.SendChatMessage(in)
}

func (s *FriendServiceServer) GetChatHistory(ctx context.Context, in *friend.GetChatHistoryRequest) (*friend.GetChatHistoryResponse, error) {
	l := logic.NewGetChatHistoryLogic(ctx, s.svcCtx)
	return l.GetChatHistory(in)
}
