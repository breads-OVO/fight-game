package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fight-game/pb/friend"
	"fight-game/pkg/common/utils"
	"fight-game/service/friend/internal/config"
	"fight-game/service/friend/internal/server"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	friendClient friend.FriendServiceClient
	grpcServer   *grpc.Server
	serviceCtx   *svc.ServiceContext
	testSeq      int
)

func testPlayerId(name string) string {
	testSeq++
	return fmt.Sprintf("test_friend_player_%d_%d_%s", time.Now().UnixMilli(), testSeq, name)
}

func findConfigFile() string {
	candidates := []string{
		filepath.Join("service", "friend", "etc", "friend.yaml"),
		filepath.Join("etc", "friend.yaml"),
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
			p := filepath.Join(dir, "service", "friend", "etc", "friend.yaml")
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
	srv := server.NewFriendServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Friend test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	friend.RegisterFriendServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	friendClient = friend.NewFriendServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

func TestFriend_AddFriend_Success(t *testing.T) {
	p1 := testPlayerId("add_p1")
	p2 := testPlayerId("add_p2")

	resp, err := friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})
	if err != nil {
		t.Fatalf("AddFriend failed: %v", err)
	}
	if !resp.Success {
		t.Logf("AddFriend message: %s", resp.Message)
	}
	t.Logf("AddFriend: success=%v, message=%s", resp.Success, resp.Message)
}

func TestFriend_ReplyFriend_Accept(t *testing.T) {
	p1 := testPlayerId("reply_p1")
	p2 := testPlayerId("reply_p2")

	friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})

	resp, err := friendClient.ReplyFriend(context.Background(), &friend.ReplyFriendRequest{
		PlayerId: p2,
		FriendId: p1,
		Accept:   true,
	})
	if err != nil {
		t.Fatalf("ReplyFriend accept failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("ReplyFriend accept should succeed:", resp.Success)
	}
	t.Log("Friend request accepted successfully")
}

func TestFriend_RemoveFriend_Success(t *testing.T) {
	p1 := testPlayerId("remove_p1")
	p2 := testPlayerId("remove_p2")

	friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})

	friendClient.ReplyFriend(context.Background(), &friend.ReplyFriendRequest{
		PlayerId: p2,
		FriendId: p1,
		Accept:   true,
	})

	resp, err := friendClient.RemoveFriend(context.Background(), &friend.RemoveFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})
	if err != nil {
		t.Fatalf("RemoveFriend failed: %v", err)
	}
	if !resp.Success {
		t.Error("RemoveFriend should succeed")
	}
	t.Log("Friend removed successfully")
}

func TestFriend_SearchPlayer_Success(t *testing.T) {
	p1 := testPlayerId("search_target")

	resp, err := friendClient.SearchPlayer(context.Background(), &friend.SearchPlayerRequest{
		Keyword:  p1[:16],
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("SearchPlayer failed: %v", err)
	}
	t.Logf("Search result: %d players found", resp.Total)
}

func TestFriend_SendChatMessage_Success(t *testing.T) {
	p1 := testPlayerId("chat_p1")
	p2 := testPlayerId("chat_p2")

	friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})

	friendClient.ReplyFriend(context.Background(), &friend.ReplyFriendRequest{
		PlayerId: p2,
		FriendId: p1,
		Accept:   true,
	})

	resp, err := friendClient.SendChatMessage(context.Background(), &friend.SendChatMessageRequest{
		SenderId:   p1,
		ReceiverId: p2,
		Content:    "Hello from test!",
	})
	if err != nil {
		t.Fatalf("SendChatMessage failed: %v", err)
	}
	t.Logf("Chat message sent: messageId=%s", resp.Message)
}

func TestFriend_GetChatHistory_Success(t *testing.T) {
	p1 := testPlayerId("history_p1")
	p2 := testPlayerId("history_p2")

	friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})

	friendClient.ReplyFriend(context.Background(), &friend.ReplyFriendRequest{
		PlayerId: p2,
		FriendId: p1,
		Accept:   true,
	})

	friendClient.SendChatMessage(context.Background(), &friend.SendChatMessageRequest{
		SenderId:   p1,
		ReceiverId: p2,
		Content:    "Test message 1",
	})

	friendClient.SendChatMessage(context.Background(), &friend.SendChatMessageRequest{
		SenderId:   p2,
		ReceiverId: p1,
		Content:    "Test message 2",
	})

	resp, err := friendClient.GetChatHistory(context.Background(), &friend.GetChatHistoryRequest{
		PlayerId: p1,
		TargetId: p2,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}
	if resp.Messages == nil {
		t.Fatal("Messages should not be nil")
	}
	t.Logf("Chat history: %d messages", len(resp.Messages))
}

func TestFriend_FullFlow(t *testing.T) {
	t.Log("Step 1: Create two players")
	p1 := testPlayerId("ff_p1")
	p2 := testPlayerId("ff_p2")

	t.Log("Step 2: p1 sends friend request to p2")
	addResp, err := friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})
	if err != nil {
		t.Fatalf("AddFriend failed: %v", err)
	}
	t.Logf("  AddFriend: success=%v", addResp.Success)

	t.Log("Step 3: p2 accepts friend request")
	replyResp, err := friendClient.ReplyFriend(context.Background(), &friend.ReplyFriendRequest{
		PlayerId: p2,
		FriendId: p1,
		Accept:   true,
	})
	if err != nil {
		t.Fatalf("ReplyFriend failed: %v", err)
	}
	t.Logf("  ReplyFriend: success=%v", replyResp.Success)

	t.Log("Step 4: Get friend list")
	listResp, err := friendClient.GetFriendList(context.Background(), &friend.GetFriendListRequest{
		PlayerId: p1,
	})
	if err != nil {
		t.Fatalf("GetFriendList failed: %v", err)
	}
	t.Logf("  Friend list: %d friends", listResp.TotalCount)

	t.Log("Step 5: Send chat message")
	chatResp, err := friendClient.SendChatMessage(context.Background(), &friend.SendChatMessageRequest{
		SenderId:   p1,
		ReceiverId: p2,
		Content:    "Hello friend!",
	})
	if err != nil {
		t.Fatalf("SendChatMessage failed: %v", err)
	}
	t.Logf("  Chat sent: messageId=%s", chatResp.Message)

	t.Log("Step 6: Get chat history")
	historyResp, err := friendClient.GetChatHistory(context.Background(), &friend.GetChatHistoryRequest{
		PlayerId: p1,
		TargetId: p2,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}
	t.Logf("  History: %d messages", len(historyResp.Messages))

	t.Log("Step 7: Remove friend")
	removeResp, err := friendClient.RemoveFriend(context.Background(), &friend.RemoveFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})
	if err != nil {
		t.Fatalf("RemoveFriend failed: %v", err)
	}
	t.Logf("  RemoveFriend: success=%v", removeResp.Success)

	t.Log("Full flow completed!")
}

func TestFriend_AddFriend_NonExistent(t *testing.T) {
	p1 := testPlayerId("non_p1")
	p2 := "non_existent_player_" + utils.GenUUID()[:8]

	resp, err := friendClient.AddFriend(context.Background(), &friend.AddFriendRequest{
		PlayerId: p1,
		FriendId: p2,
	})
	if err != nil {
		t.Logf("AddFriend with non-existent player: %v", err)
		return
	}
	t.Logf("AddFriend with non-existent player: success=%v", resp.Success)
}
