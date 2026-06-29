package server

import (
	"context"

	"fight-game/pb/mail"
	"fight-game/service/mail/internal/logic"
	"fight-game/service/mail/internal/svc"
)

type MailServiceServer struct {
	svcCtx *svc.ServiceContext
	mail.UnimplementedMailServiceServer
}

func NewMailServiceServer(svcCtx *svc.ServiceContext) *MailServiceServer {
	return &MailServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *MailServiceServer) SendMail(ctx context.Context, in *mail.SendMailRequest) (*mail.SendMailResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.SendMail(in)
}

func (s *MailServiceServer) GetMailList(ctx context.Context, in *mail.GetMailListRequest) (*mail.GetMailListResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.GetMailList(in)
}

func (s *MailServiceServer) GetMailDetail(ctx context.Context, in *mail.GetMailDetailRequest) (*mail.GetMailDetailResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.GetMailDetail(in)
}

func (s *MailServiceServer) ReadMail(ctx context.Context, in *mail.ReadMailRequest) (*mail.ReadMailResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.ReadMail(in)
}

func (s *MailServiceServer) ClaimAttachment(ctx context.Context, in *mail.ClaimAttachmentRequest) (*mail.ClaimAttachmentResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.ClaimAttachment(in)
}

func (s *MailServiceServer) DeleteMail(ctx context.Context, in *mail.DeleteMailRequest) (*mail.DeleteMailResponse, error) {
	l := logic.NewMailLogic(ctx, s.svcCtx)
	return l.DeleteMail(in)
}
