package svc

import (
	"fight-game/pb/auth"
	"fight-game/pb/friend"
	"fight-game/pb/mail"
	"fight-game/pb/match"
	"fight-game/pb/player"
	"fight-game/service/gateway/internal/config"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/ws"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	SessionMgr   *ws.SessionManager
	Router       *router.Router
	AuthRpc      zrpc.Client
	AuthClient   auth.AuthServiceClient
	MatchRpc     zrpc.Client
	MatchClient  match.MatchServiceClient
	PlayerRpc    zrpc.Client
	PlayerClient player.PlayerServiceClient
	MailClient   mail.MailServiceClient
	FriendClient friend.FriendServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {

	svc := &ServiceContext{
		Config:     c,
		SessionMgr: ws.NewSessionManager(),
	}

	if c.AuthRpc.Etcd.Key != "" {
		svc.AuthRpc = zrpc.MustNewClient(c.AuthRpc)
		svc.AuthClient = auth.NewAuthServiceClient(svc.AuthRpc.Conn())
	}

	if c.MatchRpc.Etcd.Key != "" {
		svc.MatchRpc = zrpc.MustNewClient(c.MatchRpc)
		svc.MatchClient = match.NewMatchServiceClient(svc.MatchRpc.Conn())
	}

	if c.PlayerRpc.Etcd.Key != "" {
		svc.PlayerRpc = zrpc.MustNewClient(c.PlayerRpc)
		svc.PlayerClient = player.NewPlayerServiceClient(svc.PlayerRpc.Conn())
	}

	if c.MailRpc.Etcd.Key != "" {
		mailRpc := zrpc.MustNewClient(c.MailRpc)
		svc.MailClient = mail.NewMailServiceClient(mailRpc.Conn())
	}

	if c.FriendRpc.Etcd.Key != "" {
		friendRpc := zrpc.MustNewClient(c.FriendRpc)
		svc.FriendClient = friend.NewFriendServiceClient(friendRpc.Conn())
	}

	return svc
}
