package scanner

import (
	"context"
	"fight-game/service/match/internal/scanner/match"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// MatchScanner 定时扫描器，负责周期性触发匹配
type MatchScanner struct {
	service  *match.MatchService // 匹配服务
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration // 扫描间隔，可配置
	name     string        // 扫描器名称（entertainment / competition）
}

// NewMatchScanner 创建扫描器
func NewMatchScanner(service *match.MatchService, interval time.Duration, name string) *MatchScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &MatchScanner{
		service:  service,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		name:     name,
	}
}

// Start 启动扫描循环（非阻塞）
func (s *MatchScanner) Start() {
	go s.loop()
	logx.Infof("%s match scanner started", s.name)
}

// Stop 停止扫描循环
func (s *MatchScanner) Stop() {
	s.cancel()
	logx.Infof("%s match scanner stopped", s.name)
}

// loop 定时执行匹配
func (s *MatchScanner) loop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.service.DoMatch(s.ctx) // 调用匹配服务
		}
	}
}
