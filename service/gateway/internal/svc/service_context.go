package svc

import (
	"fight-game/service/gateway/internal/config"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/ws"
)

type ServiceContext struct {
	Config     config.Config
	SessionMgr *ws.SessionManager
	Router     *router.Router
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		SessionMgr: ws.NewSessionManager(),
	}
}
