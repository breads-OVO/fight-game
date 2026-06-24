package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateTokenPair 生成 access token 和 refresh token
func GenerateTokenPair(secret string, playerId string) (accessToken string, refreshToken string, err error) {
	claims := jwt.MapClaims{
		"playerId": playerId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"playerId": playerId,
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析 token 并返回 playerId
func ParseToken(secret string, tokenStr string) (string, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	playerId, _ := claims["playerId"].(string)
	if playerId == "" {
		return "", errors.New("playerId not found in token")
	}

	return playerId, nil
}
