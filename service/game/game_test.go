package main

import (
	"context"
	"fight-game/pb/game"
	"fight-game/pb/match/queue"
	"fight-game/service/game/internal/config"
	"fight-game/service/game/internal/server"
	"fight-game/service/game/internal/svc"
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
	gameClient game.GameServiceClient
	grpcServer *grpc.Server
	serviceCtx *svc.ServiceContext
)

const (
	roomId = "test_room"
)

var playerIds = []string{
	"test_game_player_1",
	"test_game_player_2",
}

func findConfigFile() string {
	candidates := []string{
		filepath.Join("service", "game", "etc", "game.yaml"),
		filepath.Join("etc", "game.yaml"),
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
			p := filepath.Join(dir, "service", "game", "etc", "game.yaml")
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
	srv := server.NewGameServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Game test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	game.RegisterGameServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	gameClient = game.NewGameServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

// 测试初始化游戏房间_竞技
func TestGame_CreateRoom_Competition(t *testing.T) {
	roomId := roomId
	p1 := playerIds[0]
	p2 := playerIds[1]

	resp, err := gameClient.CreateRoom(context.Background(), &game.CreateRoomRequest{
		RoomId:    roomId,
		GameType:  queue.GameType_COMPETITION,
		PlayerIds: []string{p1, p2},
		Rating:    1000,
	})
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("CreateRoom should succeed, message: %s", resp.Message)
	}
	if resp.WsAddr == "" {
		t.Error("WsAddr should not be empty")
	}
	t.Logf("Created room: id=%s, wsAddr=%s", roomId, resp.WsAddr)
}

// 测试初始化游戏房间_娱乐
func TestGame_CreateRoom_Entertainment(t *testing.T) {
	roomId := roomId
	p1 := playerIds[0]
	p2 := playerIds[1]

	resp, err := gameClient.CreateRoom(context.Background(), &game.CreateRoomRequest{
		RoomId:    roomId,
		GameType:  queue.GameType_ENTERTAINMENT,
		PlayerIds: []string{p1, p2},
		Rating:    0,
	})
	if err != nil {
		t.Fatalf("CreateRoom entertainment failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("CreateRoom should succeed")
	}
	t.Logf("Created entertainment room: id=%s", roomId)
}
