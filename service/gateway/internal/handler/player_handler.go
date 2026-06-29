package handler

import (
	"context"

	"fight-game/pb/common"
	"fight-game/pb/player"
	"fight-game/pb/player/asset"
	"fight-game/pb/player/currency"
	"fight-game/pb/player/rank"
	"fight-game/pkg/common/utils"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type PlayerHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPlayerHandler(svcCtx *svc.ServiceContext) *PlayerHandler {
	return &PlayerHandler{svcCtx: svcCtx}
}

// Routes 返回玩家相关的 WS 消息路由
func (h *PlayerHandler) Routes() map[common.WSMsgType]router.HandlerFunc {
	return map[common.WSMsgType]router.HandlerFunc{
		common.WSMsgType_MSG_PLAYER_GET_PROFILE:      h.GetProfile,
		common.WSMsgType_MSG_PLAYER_LEVEL_UP:         h.LevelUp,
		common.WSMsgType_MSG_PLAYER_CHANGE_NICKNAME:  h.ChangeNickname,
		common.WSMsgType_MSG_PLAYER_CHANGE_AVATAR:    h.ChangeAvatar,
		common.WSMsgType_MSG_PLAYER_CHANGE_SIGNATURE: h.ChangeSignature,
		common.WSMsgType_MSG_PLAYER_GET_CURRENCIES:   h.GetCurrencies,
		common.WSMsgType_MSG_PLAYER_CHANGE_CURRENCY:  h.ChangeCurrency,
		common.WSMsgType_MSG_PLAYER_GET_INVENTORY:    h.GetInventory,
		common.WSMsgType_MSG_PLAYER_ADD_ASSET:        h.AddAsset,
		common.WSMsgType_MSG_PLAYER_REMOVE_ASSET:     h.RemoveAsset,
		common.WSMsgType_MSG_PLAYER_GET_RATING:       h.GetRating,
		common.WSMsgType_MSG_PLAYER_UPDATE_RATING:    h.UpdateRating,
		common.WSMsgType_MSG_PLAYER_GET_MATCH_STATS:  h.GetMatchStats,
	}
}

// GetProfile 获取玩家信息
func (h *PlayerHandler) GetProfile(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req player.GetProfileRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.GetProfile(context.Background(), &req)
	if err != nil {
		logx.Errorf("player get profile failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// LevelUp 升级
func (h *PlayerHandler) LevelUp(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req player.LevelUpRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.LevelUp(context.Background(), &req)
	if err != nil {
		logx.Errorf("player level up failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ChangeNickname 修改昵称
func (h *PlayerHandler) ChangeNickname(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req player.UpdateNicknameRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.ChangeNickname(context.Background(), &req)
	if err != nil {
		logx.Errorf("player change nickname failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ChangeAvatar 修改头像
func (h *PlayerHandler) ChangeAvatar(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req player.UpdateAvatarRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.ChangeAvatar(context.Background(), &req)
	if err != nil {
		logx.Errorf("player change avatar failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ChangeSignature 修改签名
func (h *PlayerHandler) ChangeSignature(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req player.UpdateSignatureRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.ChangeSignature(context.Background(), &req)
	if err != nil {
		logx.Errorf("player change signature failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetCurrencies 获取货币
func (h *PlayerHandler) GetCurrencies(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req currency.GetCurrenciesRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.GetCurrencies(context.Background(), &req)
	if err != nil {
		logx.Errorf("player get currencies failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// ChangeCurrency 修改货币
func (h *PlayerHandler) ChangeCurrency(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req currency.ChangeCurrencyRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.ChangeCurrency(context.Background(), &req)
	if err != nil {
		logx.Errorf("player change currency failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetInventory 获取背包
func (h *PlayerHandler) GetInventory(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req asset.GetInventoryRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.GetInventory(context.Background(), &req)
	if err != nil {
		logx.Errorf("player get inventory failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// AddAsset 添加资产
func (h *PlayerHandler) AddAsset(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req asset.AddAssetRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.AddAsset(context.Background(), &req)
	if err != nil {
		logx.Errorf("player add asset failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// RemoveAsset 移除资产
func (h *PlayerHandler) RemoveAsset(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req asset.RemoveAssetRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.RemoveAsset(context.Background(), &req)
	if err != nil {
		logx.Errorf("player remove asset failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetRating 获取段位分
func (h *PlayerHandler) GetRating(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req rank.GetRatingRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.GetRating(context.Background(), &req)
	if err != nil {
		logx.Errorf("player get rating failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// UpdateRating 修改段位分
func (h *PlayerHandler) UpdateRating(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req rank.UpdateRatingRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.UpdateRating(context.Background(), &req)
	if err != nil {
		logx.Errorf("player update rating failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}

// GetMatchStats 获取对战统计
func (h *PlayerHandler) GetMatchStats(playerId string, msg *common.WSMessage) (*common.WSResponse, error) {
	var req rank.GetMatchStatsRequest
	if err := utils.UnpackBody(msg, &req); err != nil {
		return &common.WSResponse{Code: -1, Message: "invalid request body"}, nil
	}
	req.PlayerId = playerId

	resp, err := h.svcCtx.PlayerClient.GetMatchStats(context.Background(), &req)
	if err != nil {
		logx.Errorf("player get match stats failed: player=%s, err=%v", playerId, err)
		return &common.WSResponse{Code: -1, Message: err.Error()}, nil
	}

	data, _ := proto.Marshal(resp)
	return &common.WSResponse{Code: 0, Message: "success", Data: data}, nil
}
