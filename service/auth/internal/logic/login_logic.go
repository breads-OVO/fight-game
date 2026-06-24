package logic

import (
	"context"
	"errors"

	"fight-game/pb/auth/login"
	"fight-game/pkg/common/utils"
	"fight-game/service/auth/internal/model"
	"fight-game/service/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *login.LoginRequest) (*login.LoginResponse, error) {
	var user model.User

	switch in.Type {
	case login.LoginType_LoginType_Email:
		email := in.GetEmail().GetEmail()
		if err := l.svcCtx.DB.Where("email = ?", email).First(&user).Error; err != nil {
			return nil, errors.New("账号不存在,邮箱:" + email)
		}
	case login.LoginType_LoginType_Phone:
		phone := in.GetPhone().GetPhone()
		if err := l.svcCtx.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
			return nil, errors.New("账号不存在，手机号:" + phone)
		}
	default:
		return nil, errors.New("登录类型未知")
	}

	if user.Password != utils.Sha256Entry(in.GetEmail().GetPassword()) {
		return nil, errors.New("密码错误")
	}

	token, refreshToken, err := utils.GenerateTokenPair(l.svcCtx.Config.JwtSecret, user.Id)
	if err != nil {
		return nil, err
	}

	return &login.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		PlayerId:     user.Id,
	}, nil
}
