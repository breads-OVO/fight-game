package main

import (
	"context"
	"fight-game/pb/auth"
	"fight-game/pb/auth/login"
	"fight-game/pb/auth/register"
	"fight-game/pb/auth/token"
	"fight-game/service/auth/internal/config"
	"fight-game/service/auth/internal/server"
	"fight-game/service/auth/internal/svc"
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
	authClient auth.AuthServiceClient
	grpcServer *grpc.Server
	serviceCtx *svc.ServiceContext
)

const (
	testEmail    = "test@test.com"
	testPhone    = "13800000000"
	testPassword = "123456"
)

// findConfigFile 从多个位置查找配置文件，适配不同运行环境
func findConfigFile() string {
	candidates := []string{
		// 从项目根目录运行: go test ./service/auth/
		filepath.Join("service", "auth", "etc", "auth.yaml"),
		// 从包目录运行: go test .
		filepath.Join("etc", "auth.yaml"),
		// 从项目根目录通过 -f 传入
		*configFile,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	// 尝试通过查找 go.mod 定位项目根目录
	dir, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			p := filepath.Join(dir, "service", "auth", "etc", "auth.yaml")
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
	srv := server.NewAuthServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Auth test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	authClient = auth.NewAuthServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

// 邮箱注册 测试
func TestRegister_Email_Success(t *testing.T) {
	email := testEmail
	resp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Email,
		RegisterInfo: &register.RegisterRequest_Email{
			Email: &register.RegisterEmailRequest{
				Email:    email,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register email failed: %v", err)
	}
	if resp.PlayerId == "" {
		t.Error("PlayerId should not be empty")
	}
	if resp.Token == "" {
		t.Error("Token should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
}

// 手机号注册 测试
func TestRegister_Phone_Success(t *testing.T) {
	phone := testPhone
	resp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Phone,
		RegisterInfo: &register.RegisterRequest_Phone{
			Phone: &register.RegisterPhoneRequest{
				Phone:    phone,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register phone failed: %v", err)
	}
	if resp.PlayerId == "" {
		t.Error("PlayerId should not be empty")
	}
	if resp.Token == "" {
		t.Error("Token should not be empty")
	}
}

// Login 测试

func TestLogin_Email_Success(t *testing.T) {
	email := testEmail

	// 先注册
	regResp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Email,
		RegisterInfo: &register.RegisterRequest_Email{
			Email: &register.RegisterEmailRequest{
				Email:    email,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register for login failed: %v", err)
	}

	// 登录
	resp, err := authClient.Login(context.Background(), &login.LoginRequest{
		Type: login.LoginType_LoginType_Email,
		LoginInfo: &login.LoginRequest_Email{
			Email: &login.LoginEmailRequest{
				Email:    email,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.PlayerId != regResp.PlayerId {
		t.Errorf("PlayerId mismatch: got %q, want %q", resp.PlayerId, regResp.PlayerId)
	}
	if resp.Token == "" {
		t.Error("Token should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
}

func TestLogin_Phone_Success(t *testing.T) {
	phone := testPhone

	// 先注册
	regResp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Phone,
		RegisterInfo: &register.RegisterRequest_Phone{
			Phone: &register.RegisterPhoneRequest{
				Phone:    phone,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register for login failed: %v", err)
	}

	resp, err := authClient.Login(context.Background(), &login.LoginRequest{
		Type: login.LoginType_LoginType_Phone,
		LoginInfo: &login.LoginRequest_Phone{
			Phone: &login.LoginPhoneRequest{
				Phone:    phone,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.PlayerId != regResp.PlayerId {
		t.Errorf("PlayerId mismatch: got %q, want %q", resp.PlayerId, regResp.PlayerId)
	}
}

// VerifyToken 测试

func TestVerifyToken_Valid(t *testing.T) {
	email := testEmail

	regResp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Email,
		RegisterInfo: &register.RegisterRequest_Email{
			Email: &register.RegisterEmailRequest{
				Email:    email,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: regResp.Token,
	})
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if !resp.Valid {
		t.Error("Token should be valid")
	}
	if resp.PlayerId != regResp.PlayerId {
		t.Errorf("PlayerId mismatch: got %q, want %q", resp.PlayerId, regResp.PlayerId)
	}
}

func TestVerifyToken_Invalid(t *testing.T) {
	resp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: "invalid_token_string",
	})
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if resp.Valid {
		t.Error("Invalid token should not be valid")
	}
	if resp.PlayerId != "" {
		t.Errorf("PlayerId should be empty for invalid token, got %q", resp.PlayerId)
	}
}

func TestVerifyToken_Empty(t *testing.T) {
	resp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: "",
	})
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if resp.Valid {
		t.Error("Empty token should not be valid")
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	// 用错误密钥签发的 token
	malformedToken := "eyJhbGciOiJIUzI1NiJ9.eyJwbGF5ZXJJZCI6IjEyMyJ9.dGVzdA"

	resp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: malformedToken,
	})
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if resp.Valid {
		t.Error("Malformed token should not be valid")
	}
}

// RefreshToken 测试
func TestRefreshToken_Success(t *testing.T) {
	email := testEmail

	regResp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Email,
		RegisterInfo: &register.RegisterRequest_Email{
			Email: &register.RegisterEmailRequest{
				Email:    email,
				Password: testPassword,
			},
		},
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resp, err := authClient.RefreshToken(context.Background(), &token.RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("New token should not be empty")
	}
	if resp.Token == regResp.Token {
		t.Error("New token should be different from original")
	}

	// 验证新 token 有效
	verifyResp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: resp.Token,
	})
	if err != nil {
		t.Fatalf("Verify refreshed token failed: %v", err)
	}
	if !verifyResp.Valid {
		t.Error("Refreshed token should be valid")
	}
	if verifyResp.PlayerId != regResp.PlayerId {
		t.Errorf("PlayerId mismatch: got %q, want %q", verifyResp.PlayerId, regResp.PlayerId)
	}
}

// 完整业务流程测试
func TestAuth_FullFlow(t *testing.T) {
	email := testEmail
	password := testPassword

	// 1. 注册
	regResp, err := authClient.Register(context.Background(), &register.RegisterRequest{
		Type: register.RegisterType_RegisterType_Email,
		RegisterInfo: &register.RegisterRequest_Email{
			Email: &register.RegisterEmailRequest{
				Email:    email,
				Password: password,
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 1 Register failed: %v", err)
	}
	playerId := regResp.PlayerId
	t.Logf("注册成功: playerId=%s", playerId)

	// 2. 验证 Token
	verifyResp, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: regResp.Token,
	})
	if err != nil {
		t.Fatalf("Step 2 VerifyToken failed: %v", err)
	}
	if !verifyResp.Valid || verifyResp.PlayerId != playerId {
		t.Fatalf("Token verification failed: valid=%v, playerId=%q", verifyResp.Valid, verifyResp.PlayerId)
	}
	t.Log("Token 验证通过")

	// 3. 刷新 Token
	refreshResp, err := authClient.RefreshToken(context.Background(), &token.RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("Step 3 RefreshToken failed: %v", err)
	}
	t.Log("Token 刷新成功")

	// 4. 验证新 Token
	verifyResp2, err := authClient.VerifyToken(context.Background(), &token.VerifyRequest{
		Token: refreshResp.Token,
	})
	if err != nil {
		t.Fatalf("Step 4 Verify new token failed: %v", err)
	}
	if !verifyResp2.Valid || verifyResp2.PlayerId != playerId {
		t.Fatalf("New token verification failed")
	}
	t.Log("新 Token 验证通过")

	// 5. 退出登录后再次登录（验证密码正确性）
	loginResp, err := authClient.Login(context.Background(), &login.LoginRequest{
		Type: login.LoginType_LoginType_Email,
		LoginInfo: &login.LoginRequest_Email{
			Email: &login.LoginEmailRequest{
				Email:    email,
				Password: password,
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 5 Login failed: %v", err)
	}
	if loginResp.PlayerId != playerId {
		t.Fatalf("Login returned wrong playerId: got %q, want %q", loginResp.PlayerId, playerId)
	}
	t.Log("登录成功，完整流程验证通过")
}
