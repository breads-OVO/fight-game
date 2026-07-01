package main

import (
	"context"
	"fight-game/pb/mail"
	"fight-game/service/mail/internal/config"
	"fight-game/service/mail/internal/server"
	"fight-game/service/mail/internal/svc"
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
	mailClient mail.MailServiceClient
	grpcServer *grpc.Server
	serviceCtx *svc.ServiceContext
)

var playerIds = []string{"player_1", "player_2"}

func findConfigFile() string {
	candidates := []string{
		filepath.Join("service", "mail", "etc", "mail.yaml"),
		filepath.Join("etc", "mail.yaml"),
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
			p := filepath.Join(dir, "service", "mail", "etc", "mail.yaml")
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
	srv := server.NewMailServiceServer(serviceCtx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Mail test server listening on 127.0.0.1:%d\n", port)

	grpcServer = grpc.NewServer()
	mail.RegisterMailServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial: %v\n", err)
		os.Exit(1)
	}
	mailClient = mail.NewMailServiceClient(conn)

	code := m.Run()

	grpcServer.Stop()
	conn.Close()
	os.Exit(code)
}

func TestMail_SendMail_Success(t *testing.T) {
	pid := playerIds[0]

	resp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "欢迎加入战斗游戏",
		Content:    "感谢您注册战斗游戏，祝您游戏愉快！",
		MailType:   mail.MailType_SYSTEM,
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}
	if resp.MailId == "" {
		t.Error("MailId should not be empty")
	}
	t.Logf("Mail sent: mailId=%s", resp.MailId)
}

func TestMail_SendMail_WithAttachments(t *testing.T) {
	pid := playerIds[0]

	resp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "新手礼包",
		Content:    "赠送您新手礼包一份",
		MailType:   mail.MailType_SYSTEM,
		Attachments: []*mail.Attachment{
			{
				Type:   "currency",
				Id:     "gold",
				Amount: 1000,
			},
			{
				Type:      "asset",
				Id:        "char_1",
				Amount:    1,
				AssetType: 1,
			},
		},
	})
	if err != nil {
		t.Fatalf("SendMail with attachments failed: %v", err)
	}
	t.Logf("Mail with attachments sent: mailId=%s", resp.MailId)
}

func TestMail_GetMailList_Success(t *testing.T) {
	pid := playerIds[0]

	mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "测试邮件1",
		Content:    "内容1",
		MailType:   mail.MailType_SYSTEM,
	})

	mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "测试邮件2",
		Content:    "内容2",
		MailType:   mail.MailType_SYSTEM,
	})

	resp, err := mailClient.GetMailList(context.Background(), &mail.GetMailListRequest{
		PlayerId: pid,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetMailList failed: %v", err)
	}
	if resp.Mails == nil {
		t.Fatal("Mails should not be nil")
	}
	if resp.Total < 2 {
		t.Errorf("Expected at least 2 mails, got %d", resp.Total)
	}
	t.Logf("Mail list: %d mails", resp.Total)
}

func TestMail_GetMailList_Empty(t *testing.T) {
	pid := playerIds[0]

	resp, err := mailClient.GetMailList(context.Background(), &mail.GetMailListRequest{
		PlayerId: pid,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetMailList failed: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Expected 0 mails, got %d", resp.Total)
	}
	t.Log("Empty mail list returned successfully")
}

func TestMail_GetMailDetail_Success(t *testing.T) {
	pid := playerIds[0]

	sendResp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "详情测试邮件",
		Content:    "这是一封用于测试详情的邮件",
		MailType:   mail.MailType_SYSTEM,
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	resp, err := mailClient.GetMailDetail(context.Background(), &mail.GetMailDetailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("GetMailDetail failed: %v", err)
	}
	if resp.Mail == nil {
		t.Fatal("MailDetail should not be nil")
	}
	if resp.Mail.MailId != sendResp.MailId {
		t.Errorf("MailId mismatch: got %q, want %q", resp.Mail.MailId, sendResp.MailId)
	}
	if resp.Mail.Title != "详情测试邮件" {
		t.Errorf("Title mismatch: got %q", resp.Mail.Title)
	}
	t.Logf("Mail detail: title=%s", resp.Mail.Title)
}

func TestMail_ReadMail_Success(t *testing.T) {
	pid := playerIds[0]

	sendResp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "待读邮件",
		Content:    "请阅读此邮件",
		MailType:   mail.MailType_SYSTEM,
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	resp, err := mailClient.ReadMail(context.Background(), &mail.ReadMailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("ReadMail failed: %v", err)
	}
	if !resp.Success {
		t.Error("ReadMail should succeed")
	}
	t.Log("Mail marked as read successfully")
}

func TestMail_ClaimAttachment_Success(t *testing.T) {
	pid := playerIds[0]

	sendResp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "可领取邮件",
		Content:    "点击领取附件",
		MailType:   mail.MailType_SYSTEM,
		Attachments: []*mail.Attachment{
			{
				Type:   "currency",
				Id:     "gold",
				Amount: 500,
			},
		},
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	resp, err := mailClient.ClaimAttachment(context.Background(), &mail.ClaimAttachmentRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("ClaimAttachment failed: %v", err)
	}
	if !resp.Success {
		t.Error("ClaimAttachment should succeed")
	}
	t.Log("Attachment claimed successfully")
}

func TestMail_DeleteMail_Success(t *testing.T) {
	pid := playerIds[0]

	sendResp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "待删除邮件",
		Content:    "这封邮件将被删除",
		MailType:   mail.MailType_SYSTEM,
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	resp, err := mailClient.DeleteMail(context.Background(), &mail.DeleteMailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("DeleteMail failed: %v", err)
	}
	if !resp.Success {
		t.Error("DeleteMail should succeed")
	}
	t.Log("Mail deleted successfully")
}

func TestMail_FullFlow(t *testing.T) {
	t.Log("Step 1: Send system mail")
	pid := playerIds[1]

	sendResp, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "欢迎邮件",
		Content:    "欢迎加入战斗游戏！",
		MailType:   mail.MailType_SYSTEM,
	})
	if err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}
	t.Logf("  Mail sent: mailId=%s", sendResp.MailId)

	t.Log("Step 2: Send mail with attachments")
	sendResp2, err := mailClient.SendMail(context.Background(), &mail.SendMailRequest{
		ReceiverId: pid,
		SenderId:   "system",
		SenderName: "系统",
		Title:      "奖励邮件",
		Content:    "恭喜您获得奖励！",
		MailType:   mail.MailType_SYSTEM,
		Attachments: []*mail.Attachment{
			{
				Type:   "currency",
				Id:     "gold",
				Amount: 2000,
			},
		},
	})
	if err != nil {
		t.Fatalf("SendMail with attachments failed: %v", err)
	}
	t.Logf("  Mail with attachments sent: mailId=%s", sendResp2.MailId)

	t.Log("Step 3: Get mail list")
	listResp, err := mailClient.GetMailList(context.Background(), &mail.GetMailListRequest{
		PlayerId: pid,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetMailList failed: %v", err)
	}
	t.Logf("  Mail list: %d mails", listResp.Total)

	t.Log("Step 4: Get mail detail")
	detailResp, err := mailClient.GetMailDetail(context.Background(), &mail.GetMailDetailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("GetMailDetail failed: %v", err)
	}
	t.Logf("  Mail detail: title=%s, status=%v", detailResp.Mail.Title, detailResp.Mail.Status)

	t.Log("Step 5: Mark mail as read")
	readResp, err := mailClient.ReadMail(context.Background(), &mail.ReadMailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("ReadMail failed: %v", err)
	}
	t.Logf("  ReadMail: success=%v", readResp.Success)

	t.Log("Step 6: Claim attachment")
	claimResp, err := mailClient.ClaimAttachment(context.Background(), &mail.ClaimAttachmentRequest{
		PlayerId: pid,
		MailId:   sendResp2.MailId,
	})
	if err != nil {
		t.Fatalf("ClaimAttachment failed: %v", err)
	}
	t.Logf("  ClaimAttachment: success=%v", claimResp.Success)

	t.Log("Step 7: Delete mail")
	deleteResp, err := mailClient.DeleteMail(context.Background(), &mail.DeleteMailRequest{
		PlayerId: pid,
		MailId:   sendResp.MailId,
	})
	if err != nil {
		t.Fatalf("DeleteMail failed: %v", err)
	}
	t.Logf("  DeleteMail: success=%v", deleteResp.Success)

	t.Log("Full flow completed!")
}
