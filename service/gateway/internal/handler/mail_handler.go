package handler

import (
	"context"

	"fight-game/pb/common"
	"fight-game/pb/mail"
	"fight-game/pkg/common/utils"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type MailHandler struct {
	svcCtx *svc.ServiceContext
}

func NewMailHandler(svcCtx *svc.ServiceContext) *MailHandler {
	return &MailHandler{svcCtx: svcCtx}
}

// Routes 返回邮件相关的 WS 消息路由
func (h *MailHandler) Routes() map[common.WSMsgType]router.HandlerFunc {
	return map[common.WSMsgType]router.HandlerFunc{
		common.WSMsgType_MSG_MAIL_GET_LIST:         h.GetMailList,
		common.WSMsgType_MSG_MAIL_GET_DETAIL:       h.GetMailDetail,
		common.WSMsgType_MSG_MAIL_SEND:             h.SendMail,
		common.WSMsgType_MSG_MAIL_READ:             h.ReadMail,
		common.WSMsgType_MSG_MAIL_CLAIM_ATTACHMENT: h.ClaimAttachment,
		common.WSMsgType_MSG_MAIL_DELETE:           h.DeleteMail,
	}
}

// GetMailList 获取邮件列表
func (h *MailHandler) GetMailList(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.GetMailListRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MailClient.GetMailList(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail get list failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetMailDetail 获取邮件详情
func (h *MailHandler) GetMailDetail(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.GetMailDetailRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MailClient.GetMailDetail(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail get detail failed: player=%s, mailId=%s, err=%v", playerId, req.MailId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// SendMail 发送邮件
func (h *MailHandler) SendMail(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.SendMailRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	// 默认发送者为当前玩家
	if req.SenderId == "" {
		req.SenderId = playerId
	}

	resp, err := h.svcCtx.MailClient.SendMail(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail send failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ReadMail 标记已读
func (h *MailHandler) ReadMail(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.ReadMailRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MailClient.ReadMail(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail read failed: player=%s, mailId=%s, err=%v", playerId, req.MailId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ClaimAttachment 领取附件
func (h *MailHandler) ClaimAttachment(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.ClaimAttachmentRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MailClient.ClaimAttachment(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail claim attachment failed: player=%s, mailId=%s, err=%v", playerId, req.MailId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// DeleteMail 删除邮件
func (h *MailHandler) DeleteMail(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req mail.DeleteMailRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.MailClient.DeleteMail(context.Background(), &req)
	if err != nil {
		logx.Errorf("mail delete failed: player=%s, mailId=%s, err=%v", playerId, req.MailId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}
