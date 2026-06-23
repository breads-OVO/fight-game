package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// JwtMiddleware JWT中间件
type JwtMiddleware struct {
	secret string
}

func NewJwtMiddleware(secret string) *JwtMiddleware {
	return &JwtMiddleware{secret: secret}
}

// Verify 验证请求
func (m *JwtMiddleware) Verify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			httpx.Error(w, errors.New("missing authorization header"))
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secret), nil
		})

		if err != nil || !token.Valid {
			logx.Errorf("invalid token: %v", err)
			httpx.Error(w, errors.New("invalid token"))
			return
		}

		//token.Claims 断言为 jwt.MapClaims（即 map[string]interface{} 类型）
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpx.Error(w, errors.New("invalid token claims"))
			return
		}

		playerId, _ := claims["playerId"].(string)
		r.Header.Set("X-Player-Id", playerId)
		next(w, r)
	}
}
