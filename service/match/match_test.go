package main

import (
	"context"
	"fight-game/pb/match"
	"fight-game/pb/match/queue"
	"fight-game/service/match/internal/config"
	"fight-game/service/match/internal/server"
	"fight-game/service/match/internal/svc"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	matchClient match.MatchServiceClient
	grpcServer  *grpc.Server
	serviceCtx  *svc.ServiceContext
)

var testPlayers = []string{"test_match_player_1", "test_match_player_2"}

func findConfigFile() string {
	candidates := []string{
		filepath.Join("service", "match", "etc", "match.yaml"),
		filepath.Join("etc", "match.yaml"),
		*configFile,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}

	dir, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			p := filepath.Join(dir, "service", "match", "etc", "match.yaml")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return *configFile
}

func TestMain(m *testing.M) {
	flag.Parse()

	cfgPath := findConfigFile()
	fmt.Printf("Loading config from: %s\n", cfgPath)

	var c config.Config
	conf.MustLoad(cfgPath, &c)

	serviceCtx = svc.NewServiceContext(c)
	srv := server.NewMatchServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Match test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	match.RegisterMatchServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	matchClient = match.NewMatchServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

// 竞技匹配
func TestMatch_JoinQueue_Competition(t *testing.T) {
	pid := testPlayers[0]

	resp, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
		PlayerId: pid,
		GameType: queue.GameType_COMPETITION,
		MatchParams: &queue.MatchRequest_CompetitionMatchParams{
			CompetitionMatchParams: &queue.CompetitionMatchParams{Rating: 1000},
		},
	})
	if err != nil {
		t.Fatalf("JoinQueue competition failed: %v", err)
	}
	if resp.TicketId == "" {
		t.Error("TicketId should not be empty")
	}
	if resp.EstimatedWait < 0 {
		t.Error("EstimatedWait should be >= 0")
	}
	t.Logf("Joined queue: ticket=%s", resp.TicketId)
}

// 娱乐匹配
func TestMatch_JoinQueue_Entertainment(t *testing.T) {
	pid := testPlayers[0]

	resp, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
		PlayerId: pid,
		GameType: queue.GameType_ENTERTAINMENT,
		MatchParams: &queue.MatchRequest_EntertainmentMatchParams{
			EntertainmentMatchParams: &queue.EntertainmentMatchParams{},
		},
	})
	if err != nil {
		t.Fatalf("JoinQueue entertainment failed: %v", err)
	}
	if resp.TicketId == "" {
		t.Error("TicketId should not be empty")
	}
	t.Logf("Joined entertainment queue: ticket=%s", resp.TicketId)
}

// 获取匹配状态
func TestMatch_GetMatchStatus_Queuing(t *testing.T) {
	pid := testPlayers[0]

	joinResp, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
		PlayerId: pid,
		GameType: queue.GameType_COMPETITION,
		MatchParams: &queue.MatchRequest_CompetitionMatchParams{
			CompetitionMatchParams: &queue.CompetitionMatchParams{Rating: 1000},
		},
	})
	if err != nil {
		t.Fatalf("JoinQueue failed: %v", err)
	}

	statusResp, err := matchClient.GetMatchStatus(context.Background(), &queue.MatchStatusRequest{
		TicketId: joinResp.TicketId,
	})
	if err != nil {
		t.Fatalf("GetMatchStatus failed: %v", err)
	}
	if statusResp.Status != queue.MatchStatus_QUEUEING {
		t.Errorf("Expected QUEUEING, got %v", statusResp.Status)
	}
	t.Logf("Match status: %v", statusResp.Status)
}

// 离开队列
func TestMatch_LeaveQueue_Success(t *testing.T) {
	pid := testPlayers[0]

	joinResp, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
		PlayerId: pid,
		GameType: queue.GameType_COMPETITION,
		MatchParams: &queue.MatchRequest_CompetitionMatchParams{
			CompetitionMatchParams: &queue.CompetitionMatchParams{Rating: 1000},
		},
	})
	if err != nil {
		t.Fatalf("JoinQueue failed: %v", err)
	}

	leaveResp, err := matchClient.LeaveQueue(context.Background(), &queue.LeaveQueueRequest{
		PlayerId: pid,
		TicketId: joinResp.TicketId,
	})
	if err != nil {
		t.Fatalf("LeaveQueue failed: %v", err)
	}
	if !leaveResp.Success {
		t.Error("LeaveQueue should succeed")
	}
	t.Log("LeaveQueue successful")
}

func TestMatch_FullFlow_JoinAndLeave(t *testing.T) {
	pid := testPlayers[0]

	t.Log("Step 1: Join queue")
	joinResp, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
		PlayerId: pid,
		GameType: queue.GameType_COMPETITION,
		MatchParams: &queue.MatchRequest_CompetitionMatchParams{
			CompetitionMatchParams: &queue.CompetitionMatchParams{Rating: 1200},
		},
	})
	if err != nil {
		t.Fatalf("JoinQueue failed: %v", err)
	}
	t.Logf("  Ticket: %s", joinResp.TicketId)

	t.Log("Step 2: Check status (queuing)")
	statusResp, err := matchClient.GetMatchStatus(context.Background(), &queue.MatchStatusRequest{
		TicketId: joinResp.TicketId,
	})
	if err != nil {
		t.Fatalf("GetMatchStatus failed: %v", err)
	}
	t.Logf("  Status: %v", statusResp.Status)

	t.Log("Step 3: Leave queue")
	leaveResp, err := matchClient.LeaveQueue(context.Background(), &queue.LeaveQueueRequest{
		PlayerId: pid,
		TicketId: joinResp.TicketId,
	})
	if err != nil {
		t.Fatalf("LeaveQueue failed: %v", err)
	}
	t.Logf("  Leave success: %v", leaveResp.Success)

	t.Log("Step 4: Check status after leave")
	statusResp2, err := matchClient.GetMatchStatus(context.Background(), &queue.MatchStatusRequest{
		TicketId: joinResp.TicketId,
	})
	if err != nil {
		t.Fatalf("GetMatchStatus after leave failed: %v", err)
	}
	t.Logf("  Status after leave: %v", statusResp2.Status)

	t.Log("Full flow completed!")
}

func TestMatch_Concurrent_JoinQueue(t *testing.T) {
	errs := make(chan error, 5)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			pid := testPlayers[i]
			_, err := matchClient.JoinQueue(context.Background(), &queue.MatchRequest{
				PlayerId: pid,
				GameType: queue.GameType_COMPETITION,
				MatchParams: &queue.MatchRequest_CompetitionMatchParams{
					CompetitionMatchParams: &queue.CompetitionMatchParams{Rating: 1000 + int32(idx*50)},
				},
			})
			errs <- err
		}(i)
	}

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Errorf("Concurrent join failed: %v", err)
		}
	}
	t.Log("Concurrent JoinQueue completed successfully")
}
