package logic

import (
	"context"
	"errors"

	"fight-game/pb/auth/register"
	"fight-game/pkg/common/utils"
	"fight-game/service/auth/internal/model"
	"fight-game/service/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *register.RegisterRequest) (*register.RegisterResponse, error) {

	var email, phone, password string
	switch in.Type {
	case register.RegisterType_RegisterType_Email:
		email = in.GetEmail().GetEmail()
		password = in.GetEmail().GetPassword()
		// 检查账号是否存在
		existing := l.svcCtx.DB.Where("email = ? ", email).First(&model.User{})
		if existing.RowsAffected > 0 {
			return nil, errors.New("该邮箱已注册，邮箱:" + email)
		}
	case register.RegisterType_RegisterType_Phone:
		phone = in.GetPhone().GetPhone()
		password = in.GetPhone().GetPassword()
		// 检查账号是否存在
		existing := l.svcCtx.DB.Where("phone = ? ", phone).First(&model.User{})
		if existing.RowsAffected > 0 {
			return nil, errors.New("该手机号已注册，手机号:" + phone)
		}
	}

	// 保存用户信息
	user := &model.User{
		Password: utils.Sha256Entry(password),
		Email:    email,
		Phone:    phone,
	}
	user.Id = utils.GenUUID()
	user.CreatedAt = utils.GetNowTime()
	if err := l.svcCtx.DB.Create(user).Error; err != nil {
		return nil, err
	}

	token, refreshToken, err := utils.GenerateTokenPair(l.svcCtx.Config.JwtSecret, user.Id)
	if err != nil {
		return nil, err
	}

	return &register.RegisterResponse{
		Token:        token,
		RefreshToken: refreshToken,
		PlayerId:     user.Id,
	}, nil
}
