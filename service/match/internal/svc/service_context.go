package svc

import (
	"time"

	commonConf "fight-game/pkg/common/config"
	"fight-game/service/match/internal/config"
	"fight-game/service/match/internal/scanner"
	"fight-game/service/match/internal/scanner/handler"
	"fight-game/service/match/internal/scanner/infra"
	"fight-game/service/match/internal/scanner/match"
	"fight-game/service/match/internal/scanner/strategy"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	entertainmentQueueKey = "match:queue:entertainment" // 娱乐匹配队列
	competitionQueueKey   = "match:queue:competition"   // 竞技匹配队列
)

type ServiceContext struct {
	Config               config.Config
	DB                   *gorm.DB
	Redis                redis.UniversalClient
	EntertainmentQueue   match.MatchQueue
	CompetitionQueue     match.MatchQueue
	Repo                 match.TicketRepository
	EntertainmentSvc     *match.MatchService
	CompetitionSvc       *match.MatchService
	EntertainmentScanner *scanner.MatchScanner
	CompetitionScanner   *scanner.MatchScanner
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db := commonConf.InitMySQL(&c.MySQL)

	// 初始化 Redis
	redisClient := commonConf.InitRedis(&c.Redis)

	// 创建票存储仓储（共享）
	repo := infra.NewRedisTicketRepo(redisClient)

	// 创建各自的匹配队列
	entertainmentQueue := infra.NewRedisMatchQueue(redisClient, entertainmentQueueKey)
	competitionQueue := infra.NewRedisCompetitionQueue(redisClient, competitionQueueKey)

	// 结果处理器（后续可对接 Game 服务）
	handler := &handler.NoopResultHandler{}

	// --- 娱乐匹配（List + 简单FIFO） ---
	entertainmentStrategy := &strategy.EntertainmentStrategy{}
	entertainmentSvc := match.NewMatchService(entertainmentQueue, repo, entertainmentStrategy, handler)

	// --- 竞技匹配（ZSet + 段位分扩圈） ---
	competitionStrategy := strategy.NewCompetitiveStrategy(
		c.Match.RatingRange,
		c.Match.RatingRangeMax,
		c.Match.RatingRangeStep,
	)
	competitionSvc := match.NewMatchService(competitionQueue, repo, competitionStrategy, handler)

	// 创建各自的扫描器，间隔1秒
	entertainmentScanner := scanner.NewMatchScanner(entertainmentSvc, time.Second, "entertainment")
	competitionScanner := scanner.NewMatchScanner(competitionSvc, time.Second, "competition")

	return &ServiceContext{
		Config:               c,
		DB:                   db,
		Redis:                redisClient,
		EntertainmentQueue:   entertainmentQueue,
		CompetitionQueue:     competitionQueue,
		Repo:                 repo,
		EntertainmentSvc:     entertainmentSvc,
		CompetitionSvc:       competitionSvc,
		EntertainmentScanner: entertainmentScanner,
		CompetitionScanner:   competitionScanner,
	}
}
