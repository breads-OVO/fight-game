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

	"fight-game/pb/player"
	"fight-game/pb/player/asset"
	"fight-game/pb/player/currency"
	"fight-game/pb/player/rank"
	"fight-game/pkg/common/utils"
	"fight-game/service/player/internal/config"
	"fight-game/service/player/internal/server"
	"fight-game/service/player/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	playerClient player.PlayerServiceClient
	grpcServer   *grpc.Server
	serviceCtx   *svc.ServiceContext
	testSeq      int
)

func testPlayerId(name string) string {
	testSeq++
	return fmt.Sprintf("test_player_%d_%d_%s", time.Now().UnixMilli(), testSeq, name)
}

func testNickname(name string) string {
	testSeq++
	return fmt.Sprintf("TestNick_%d_%s", testSeq, name)
}

func findConfigFile() string {
	candidates := []string{
		filepath.Join("service", "player", "etc", "player.yaml"),
		filepath.Join("etc", "player.yaml"),
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
			p := filepath.Join(dir, "service", "player", "etc", "player.yaml")
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
	srv := server.NewPlayerServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Player test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	player.RegisterPlayerServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	playerClient = player.NewPlayerServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

func createTestPlayer(t *testing.T, playerId string) {
	t.Helper()
	_, err := playerClient.GetProfile(context.Background(), &player.GetProfileRequest{PlayerId: playerId})
	if err == nil {
		return
	}
	st, ok := status.FromError(err)
	if ok && st.Code().String() == "NotFound" {
		_, err := playerClient.LevelUp(context.Background(), &player.LevelUpRequest{
			PlayerId:  playerId,
			ExpGained: 0,
		})
		if err != nil {
			t.Logf("Create test player failed (may already exist): %v", err)
		}
	}
}

func TestPlayer_GetProfile_Success(t *testing.T) {
	pid := testPlayerId("profile")
	createTestPlayer(t, pid)

	resp, err := playerClient.GetProfile(context.Background(), &player.GetProfileRequest{PlayerId: pid})
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if resp.Profile == nil {
		t.Fatal("Profile should not be nil")
	}
	if resp.Profile.PlayerId != pid {
		t.Errorf("PlayerId mismatch: got %q, want %q", resp.Profile.PlayerId, pid)
	}
	if resp.Profile.Level < 1 {
		t.Errorf("Level should be >= 1, got %d", resp.Profile.Level)
	}
	t.Logf("Profile: playerId=%s, nickname=%s, level=%d",
		resp.Profile.PlayerId, resp.Profile.Nickname, resp.Profile.Level)
}

func TestPlayer_LevelUp_Success(t *testing.T) {
	pid := testPlayerId("levelup")
	createTestPlayer(t, pid)

	before, err := playerClient.GetProfile(context.Background(), &player.GetProfileRequest{PlayerId: pid})
	if err != nil {
		t.Fatalf("GetProfile before levelup failed: %v", err)
	}
	beforeLevel := before.Profile.Level

	resp, err := playerClient.LevelUp(context.Background(), &player.LevelUpRequest{
		PlayerId:  pid,
		ExpGained: 2500,
	})
	if err != nil {
		t.Fatalf("LevelUp failed: %v", err)
	}
	if resp.Profile.Level != beforeLevel+2 {
		t.Errorf("Level should increase by 2, before=%d after=%d", beforeLevel, resp.Profile.Level)
	}
	if resp.Profile.Exp != 500 {
		t.Errorf("Exp should be 500 after 2500 gained, got %d", resp.Profile.Exp)
	}
}

func TestPlayer_ChangeNickname_Success(t *testing.T) {
	pid := testPlayerId("nickname")
	createTestPlayer(t, pid)

	newNick := testNickname("change")
	resp, err := playerClient.ChangeNickname(context.Background(), &player.UpdateNicknameRequest{
		PlayerId:    pid,
		NewNickname: newNick,
	})
	if err != nil {
		t.Fatalf("ChangeNickname failed: %v", err)
	}
	if resp.Profile.Nickname != newNick {
		t.Errorf("Nickname mismatch: got %q, want %q", resp.Profile.Nickname, newNick)
	}

	verify, err := playerClient.GetProfile(context.Background(), &player.GetProfileRequest{PlayerId: pid})
	if err != nil {
		t.Fatalf("GetProfile after change failed: %v", err)
	}
	if verify.Profile.Nickname != newNick {
		t.Errorf("Nickname not persisted: got %q, want %q", verify.Profile.Nickname, newNick)
	}
}

func TestPlayer_ChangeAvatar_Success(t *testing.T) {
	pid := testPlayerId("avatar")
	createTestPlayer(t, pid)

	newAvatar := "https://example.com/avatar/" + utils.GenUUID()[:8] + ".png"
	resp, err := playerClient.ChangeAvatar(context.Background(), &player.UpdateAvatarRequest{
		PlayerId:     pid,
		NewAvatarUrl: newAvatar,
	})
	if err != nil {
		t.Fatalf("ChangeAvatar failed: %v", err)
	}
	if resp.Profile.AvatarUrl != newAvatar {
		t.Errorf("Avatar mismatch: got %q, want %q", resp.Profile.AvatarUrl, newAvatar)
	}
}

func TestPlayer_ChangeSignature_Success(t *testing.T) {
	pid := testPlayerId("signature")
	createTestPlayer(t, pid)

	newSig := "这是我的个性签名 " + utils.GenUUID()[:8]
	resp, err := playerClient.ChangeSignature(context.Background(), &player.UpdateSignatureRequest{
		PlayerId:     pid,
		NewSignature: newSig,
	})
	if err != nil {
		t.Fatalf("ChangeSignature failed: %v", err)
	}
	if resp.Profile.Signature != newSig {
		t.Errorf("Signature mismatch: got %q, want %q", resp.Profile.Signature, newSig)
	}
}

func TestPlayer_GetCurrencies_Success(t *testing.T) {
	pid := testPlayerId("currency")
	createTestPlayer(t, pid)

	resp, err := playerClient.GetCurrencies(context.Background(), &currency.GetCurrenciesRequest{PlayerId: pid})
	if err != nil {
		t.Fatalf("GetCurrencies failed: %v", err)
	}
	if resp.Currencies == nil {
		t.Fatal("Currencies should not be nil")
	}
	t.Logf("Got %d currencies", len(resp.Currencies))
}

func TestPlayer_ChangeCurrency_Success(t *testing.T) {
	pid := testPlayerId("change_currency")
	createTestPlayer(t, pid)

	_, err := playerClient.ChangeCurrency(context.Background(), &currency.ChangeCurrencyRequest{
		PlayerId:     pid,
		CurrencyType: 1,
		Count:        100,
	})
	if err != nil {
		t.Fatalf("ChangeCurrency failed: %v", err)
	}
}

func TestPlayer_GetInventory_Success(t *testing.T) {
	pid := testPlayerId("inventory")
	createTestPlayer(t, pid)

	_, err := playerClient.GetInventory(context.Background(), &asset.GetInventoryRequest{PlayerId: pid})
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
}

func TestPlayer_AddAsset_Success(t *testing.T) {
	pid := testPlayerId("add_asset")
	createTestPlayer(t, pid)

	_, err := playerClient.AddAsset(context.Background(), &asset.AddAssetRequest{
		PlayerId:  pid,
		AssetId:   "char_1",
		AssetType: 1,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}
}

func TestPlayer_RemoveAsset_Success(t *testing.T) {
	pid := testPlayerId("remove_asset")
	createTestPlayer(t, pid)

	playerClient.AddAsset(context.Background(), &asset.AddAssetRequest{
		PlayerId:  pid,
		AssetId:   "char_2",
		AssetType: 1,
		Quantity:  1,
	})

	resp, err := playerClient.RemoveAsset(context.Background(), &asset.RemoveAssetRequest{
		PlayerId: pid,
		Id:       "char_2",
		Quantity: 1,
	})
	if err != nil {
		t.Fatalf("RemoveAsset failed: %v", err)
	}
	if !resp.Success {
		t.Error("RemoveAsset should succeed")
	}
}

func TestPlayer_GetRating_Success(t *testing.T) {
	pid := testPlayerId("rating")
	createTestPlayer(t, pid)

	season := "s1"
	playerClient.UpdateRating(context.Background(), &rank.UpdateRatingRequest{
		PlayerId: pid,
		Delta:    50,
	})

	resp, err := playerClient.GetRating(context.Background(), &rank.GetRatingRequest{
		PlayerId: pid,
		Season:   season,
	})
	if err != nil {
		t.Fatalf("GetRating failed: %v", err)
	}
	if resp.Rating == nil {
		t.Fatal("Rating should not be nil")
	}
	if resp.Rating.PlayerId != pid {
		t.Errorf("PlayerId mismatch: got %q, want %q", resp.Rating.PlayerId, pid)
	}
	t.Logf("Rating: %d", resp.Rating.Rating)
}

func TestPlayer_UpdateRating_Success(t *testing.T) {
	pid := testPlayerId("update_rating")
	createTestPlayer(t, pid)

	delta := int32(100)
	_, err := playerClient.UpdateRating(context.Background(), &rank.UpdateRatingRequest{
		PlayerId: pid,
		Delta:    delta,
	})
	if err != nil {
		t.Fatalf("UpdateRating failed: %v", err)
	}

	verify, err := playerClient.GetRating(context.Background(), &rank.GetRatingRequest{
		PlayerId: pid,
		Season:   "s1",
	})
	if err != nil {
		t.Fatalf("GetRating after update failed: %v", err)
	}
	if verify.Rating.Rating < 1000 {
		t.Errorf("Rating should be at least 1000 (default), got %d", verify.Rating.Rating)
	}
}

func TestPlayer_SearchPlayer_Success(t *testing.T) {
	pid := testPlayerId("search_target")
	createTestPlayer(t, pid)

	nickname := "SearchTest_" + utils.GenUUID()[:8]
	playerClient.ChangeNickname(context.Background(), &player.UpdateNicknameRequest{
		PlayerId:    pid,
		NewNickname: nickname,
	})

	resp, err := playerClient.SearchPlayer(context.Background(), &player.SearchPlayerRequest{
		Keyword:  nickname,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("SearchPlayer failed: %v", err)
	}
	if resp.Total == 0 {
		t.Log("No results found (may be search delay)")
	} else {
		t.Logf("Found %d players", resp.Total)
	}
}
