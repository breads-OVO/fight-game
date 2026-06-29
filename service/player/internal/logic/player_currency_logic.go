package logic

import (
	"context"
	"errors"
	"fight-game/pkg/common/utils"

	"fight-game/pb/player/currency"
	"fight-game/service/player/internal/model"
	"fight-game/service/player/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const maxRetries = 3 // 最大重试次数

type PlayerCurrencyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlayerCurrencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerCurrencyLogic {
	return &PlayerCurrencyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCurrencies 获取玩家所有货币
func (l *PlayerCurrencyLogic) GetCurrencies(in *currency.GetCurrenciesRequest) (*currency.GetCurrenciesResponse, error) {
	playerId := in.GetPlayerId()

	var records []model.PlayerCurrency
	result := l.svcCtx.DB.Where("player_id = ?", playerId).Find(&records)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "query currencies failed: %v", result.Error)
	}

	infos := make([]*currency.CurrencyInfo, 0, len(records))
	for _, r := range records {
		infos = append(infos, &currency.CurrencyInfo{
			CurrencyType: currency.CurrencyType(r.CurrencyType),
			Count:        r.Amount,
		})
	}

	return &currency.GetCurrenciesResponse{
		Currencies: infos,
	}, nil
}

// ChangeCurrency 修改货币（乐观锁，最多重试3次）
func (l *PlayerCurrencyLogic) ChangeCurrency(in *currency.ChangeCurrencyRequest) (*currency.ChangeCurrencyResponse, error) {
	playerId := in.GetPlayerId()
	ct := int16(in.GetCurrencyType())
	delta := in.GetCount()
	isAdd := in.GetChangeType()

	for i := 0; i < maxRetries; i++ {
		// 1. 读取当前记录（含版本号）
		var record model.PlayerCurrency
		result := l.svcCtx.DB.Where("player_id = ? AND currency_type = ?", playerId, ct).First(&record)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.Internal, "query currency failed: %v", result.Error)
			}

			// 记录不存在：只有增加操作才允许自动创建
			if !isAdd {
				return nil, status.Errorf(codes.FailedPrecondition,
					"currency type %d not found for player %s, cannot deduct", ct, playerId)
			}

			newRecord := &model.PlayerCurrency{
				PlayerId:     playerId,
				CurrencyType: ct,
				Amount:       delta,
				Version:      1,
			}
			newRecord.Base.Id = utils.GenUUID()
			newRecord.Base.CreatedAt = utils.GetNowTime()
			if err := l.svcCtx.DB.Create(newRecord).Error; err != nil {
				return nil, status.Errorf(codes.Internal, "create currency failed: %v", err)
			}

			logx.Infof("currency created: player=%s type=%d amount=%d reason=%s",
				playerId, ct, delta, in.GetReason())
			return &currency.ChangeCurrencyResponse{
				Currency: &currency.CurrencyInfo{
					CurrencyType: currency.CurrencyType(ct),
					Count:        delta,
				},
			}, nil
		}

		// 2. 计算新值
		newAmount := record.Amount
		if isAdd {
			newAmount += delta
		} else {
			if record.Amount < delta {
				return nil, status.Errorf(codes.FailedPrecondition,
					"insufficient currency type %d: have %d, need %d", ct, record.Amount, delta)
			}
			newAmount -= delta
		}

		// 3. 乐观锁更新：WHERE version = oldVersion
		oldVersion := record.Version
		result = l.svcCtx.DB.Model(&model.PlayerCurrency{}).
			Where("id = ? AND version = ?", record.Id, oldVersion).
			Updates(map[string]interface{}{
				"amount":     newAmount,
				"version":    oldVersion + 1,
				"updated_at": utils.GetNowTime(),
			})

		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "update currency failed: %v", result.Error)
		}

		if result.RowsAffected > 0 {
			logx.Infof("currency changed: player=%s type=%d amount:%d->%d delta=%d add=%v reason=%s",
				playerId, ct, record.Amount, newAmount, delta, isAdd, in.GetReason())
			return &currency.ChangeCurrencyResponse{
				Currency: &currency.CurrencyInfo{
					CurrencyType: currency.CurrencyType(ct),
					Count:        newAmount,
				},
			}, nil
		}

		// 版本冲突，重试
		logx.Infof("optimistic lock conflict, retry %d/%d: player=%s type=%d",
			i+1, maxRetries, playerId, ct)
	}

	return nil, status.Errorf(codes.Aborted,
		"currency update conflict after %d retries: player=%s type=%d", maxRetries, playerId, ct)
}
