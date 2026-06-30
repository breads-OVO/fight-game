package logic

import (
	"context"
	"errors"
	"time"

	"fight-game/pb/player"
	"fight-game/service/player/internal/model"
	"fight-game/service/player/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PlayerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerLogic {
	return &PlayerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetProfile 获取玩家基本信息
func (l *PlayerLogic) GetProfile(in *player.GetProfileRequest) (*player.GetProfileResponse, error) {
	playerId := in.GetPlayerId()

	var p model.Player
	result := l.svcCtx.DB.Where("player_id = ?", playerId).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player %s not found", playerId)
		}
		return nil, status.Errorf(codes.Internal, "query player failed: %v", result.Error)
	}

	return &player.GetProfileResponse{
		Profile: modelToProfile(&p),
	}, nil
}

// LevelUp 升级（增加经验，检查是否升级）
func (l *PlayerLogic) LevelUp(in *player.LevelUpRequest) (*player.LevelUpResponse, error) {
	playerId := in.GetPlayerId()
	expGained := in.GetExpGained()

	var p model.Player
	result := l.svcCtx.DB.Where("player_id = ?", playerId).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player %s not found", playerId)
		}
		return nil, status.Errorf(codes.Internal, "query player failed: %v", result.Error)
	}

	// 简单升级逻辑：每满1000经验升1级
	p.Exp += int64(expGained)
	for p.Exp >= 1000 {
		p.Exp -= 1000
		p.Level++
	}
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.DB.Save(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "save player level failed: %v", err)
	}

	logx.Infof("player level up: player=%s level=%d exp=%d expGained=%d",
		playerId, p.Level, p.Exp, expGained)

	return &player.LevelUpResponse{
		Profile: modelToProfile(&p),
	}, nil
}

// ChangeNickname 修改昵称
func (l *PlayerLogic) ChangeNickname(in *player.UpdateNicknameRequest) (*player.UpdateNicknameResponse, error) {
	playerId := in.GetPlayerId()
	newNickname := in.GetNewNickname()

	var p model.Player
	result := l.svcCtx.DB.Where("player_id = ?", playerId).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player %s not found", playerId)
		}
		return nil, status.Errorf(codes.Internal, "query player failed: %v", result.Error)
	}

	p.Nickname = newNickname
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.DB.Save(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "update nickname failed: %v", err)
	}

	return &player.UpdateNicknameResponse{
		Profile: modelToProfile(&p),
	}, nil
}

// ChangeAvatar 修改头像
func (l *PlayerLogic) ChangeAvatar(in *player.UpdateAvatarRequest) (*player.UpdateAvatarResponse, error) {
	playerId := in.GetPlayerId()
	newAvatarUrl := in.GetNewAvatarUrl()

	var p model.Player
	result := l.svcCtx.DB.Where("player_id = ?", playerId).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player %s not found", playerId)
		}
		return nil, status.Errorf(codes.Internal, "query player failed: %v", result.Error)
	}

	p.AvatarURL = newAvatarUrl
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.DB.Save(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "update avatar failed: %v", err)
	}

	return &player.UpdateAvatarResponse{
		Profile: modelToProfile(&p),
	}, nil
}

// ChangeSignature 修改个性签名
func (l *PlayerLogic) ChangeSignature(in *player.UpdateSignatureRequest) (*player.UpdateSignatureResponse, error) {
	playerId := in.GetPlayerId()
	newSignature := in.GetNewSignature()

	var p model.Player
	result := l.svcCtx.DB.Where("player_id = ?", playerId).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "player %s not found", playerId)
		}
		return nil, status.Errorf(codes.Internal, "query player failed: %v", result.Error)
	}

	p.Signature = newSignature
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.DB.Save(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "update signature failed: %v", err)
	}

	return &player.UpdateSignatureResponse{
		Profile: modelToProfile(&p),
	}, nil
}

// SearchPlayer 搜索玩家（按昵称或ID模糊匹配）
func (l *PlayerLogic) SearchPlayer(in *player.SearchPlayerRequest) (*player.SearchPlayerResponse, error) {
	keyword := in.GetKeyword()
	page := int(in.GetPage())
	pageSize := int(in.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var total int64
	var players []model.Player

	// 按昵称或ID模糊搜索
	likePattern := "%" + keyword + "%"
	query := l.svcCtx.DB.Model(&model.Player{}).
		Where("nickname LIKE ? OR player_id LIKE ?", likePattern, likePattern)

	if err := query.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "count players failed: %v", err)
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&players).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "search players failed: %v", err)
	}

	profiles := make([]*player.PlayerProfile, 0, len(players))
	for i := range players {
		profiles = append(profiles, modelToProfile(&players[i]))
	}

	return &player.SearchPlayerResponse{
		Players: profiles,
		Total:   int32(total),
	}, nil
}

// modelToProfile model -> proto 转换
func modelToProfile(p *model.Player) *player.PlayerProfile {
	return &player.PlayerProfile{
		PlayerId:  p.UserID,
		Nickname:  p.Nickname,
		Level:     p.Level,
		Exp:       p.Exp,
		AvatarUrl: p.AvatarURL,
		Signature: p.Signature,
		CreatedAt: p.CreatedAt.UnixMilli(),
		UpdatedAt: p.UpdatedAt.UnixMilli(),
	}
}
