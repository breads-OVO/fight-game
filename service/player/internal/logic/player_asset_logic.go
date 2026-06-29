package logic

import (
	"context"
	"errors"
	"fight-game/pkg/common/utils"
	"time"

	"fight-game/pb/player/asset"
	"fight-game/service/player/internal/model"
	"fight-game/service/player/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PlayerAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlayerAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerAssetLogic {
	return &PlayerAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetInventory 获取玩家全部资产
func (l *PlayerAssetLogic) GetInventory(in *asset.GetInventoryRequest) (*asset.GetInventoryResponse, error) {
	playerId := in.GetPlayerId()

	var records []model.PlayerAsset
	result := l.svcCtx.DB.Where("player_id = ?", playerId).Find(&records)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "query inventory failed: %v", result.Error)
	}

	infos := make([]*asset.AssetInfo, 0, len(records))
	for _, r := range records {
		infos = append(infos, modelToAssetInfo(&r))
	}

	return &asset.GetInventoryResponse{
		Assets: infos,
	}, nil
}

// AddAsset 添加资产（assetId + assetType + expireAt 相同时叠加数量）
func (l *PlayerAssetLogic) AddAsset(in *asset.AddAssetRequest) (*asset.AddAssetResponse, error) {
	playerId := in.GetPlayerId()
	assetId := in.GetAssetId()
	assetType := int16(in.GetAssetType())
	quantity := in.GetQuantity()
	expireAt := in.GetExpireAt()
	now := utils.GetNowTime()

	// 尝试查找已存在的相同资产（相同 assetId + assetType + 过期时间）
	var existing model.PlayerAsset
	result := l.svcCtx.DB.Where("player_id = ? AND asset_id = ? AND asset_type = ? AND expire_at = ?",
		playerId, assetId, assetType, expireAt).First(&existing)

	if result.Error == nil {
		// 已存在，叠加数量
		existing.Quantity += quantity
		existing.UpdatedAt = now
		if err := l.svcCtx.DB.Save(&existing).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "update asset quantity failed: %v", err)
		}
		logx.Infof("asset quantity increased: player=%s assetId=%s type=%d quantity=%d->%d",
			playerId, assetId, assetType, existing.Quantity-quantity, existing.Quantity)
		return &asset.AddAssetResponse{
			Item: modelToAssetInfo(&existing),
		}, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "query existing asset failed: %v", result.Error)
	}

	// 不存在，创建新记录
	newAsset := &model.PlayerAsset{
		PlayerId:  playerId,
		AssetId:   assetId,
		AssetType: assetType,
		Quantity:  quantity,
		Status:    int16(asset.AssetStatus_NORMAL),
		ExpireAt:  expireAt,
	}
	newAsset.Id = uuid.New().String()
	newAsset.CreatedAt = now

	if err := l.svcCtx.DB.Create(newAsset).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "create asset failed: %v", err)
	}

	logx.Infof("asset created: player=%s assetId=%s type=%d quantity=%d expireAt=%d",
		playerId, assetId, assetType, quantity, expireAt)
	return &asset.AddAssetResponse{
		Item: modelToAssetInfo(newAsset),
	}, nil
}

// RemoveAsset 移除资产（按背包内id扣减，数量归零则删除记录）
func (l *PlayerAssetLogic) RemoveAsset(in *asset.RemoveAssetRequest) (*asset.RemoveAssetResponse, error) {
	playerId := in.GetPlayerId()
	id := in.GetId()
	quantity := in.GetQuantity()

	// 查找资产（按背包内唯一ID）
	var record model.PlayerAsset
	result := l.svcCtx.DB.Where("id = ? AND player_id = ?", id, playerId).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "asset %s not found for player %s", id, playerId)
		}
		return nil, status.Errorf(codes.Internal, "query asset failed: %v", result.Error)
	}

	if record.Quantity < quantity {
		return nil, status.Errorf(codes.FailedPrecondition,
			"insufficient asset %s quantity: have %d, need %d", id, record.Quantity, quantity)
	}

	if quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "remove quantity must be positive")
	}

	record.Quantity -= quantity
	record.UpdatedAt = utils.GetNowTime()
	if record.Quantity <= 0 {
		// 数量归零，删除记录
		if err := l.svcCtx.DB.Delete(&record).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "delete asset failed: %v", err)
		}
		logx.Infof("asset deleted (quantity zero): player=%s assetId=%s type=%d id=%s",
			playerId, record.AssetId, record.AssetType, id)
	} else {
		// 更新数量
		if err := l.svcCtx.DB.Save(&record).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "update asset quantity failed: %v", err)
		}
		logx.Infof("asset quantity decreased: player=%s assetId=%s type=%d quantity=%d->%d",
			playerId, record.AssetId, record.AssetType, record.Quantity+quantity, record.Quantity)
	}

	return &asset.RemoveAssetResponse{
		Success: true,
	}, nil
}

// modelToAssetInfo model -> proto 转换
func modelToAssetInfo(r *model.PlayerAsset) *asset.AssetInfo {
	info := &asset.AssetInfo{
		Id:        r.Id,
		AssetId:   r.AssetId,
		AssetType: asset.AssetType(r.AssetType),
		Status:    asset.AssetStatus(r.Status),
		Quantity:  r.Quantity,
		ExpireAt:  r.ExpireAt,
		CreatedAt: r.CreatedAt.UnixMilli(),
	}

	// 检查是否过期（非永久且已过期 → 标记为过期状态）
	if r.ExpireAt > 0 && time.Now().UnixMilli() > r.ExpireAt {
		info.Status = asset.AssetStatus_TEMP
	}

	return info
}
