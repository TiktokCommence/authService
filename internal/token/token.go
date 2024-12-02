package token

import (
	"errors"
	"fmt"
	"github.com/TiktokCommence/authService/internal/biz"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"time"
)

var (
	_ biz.TokenGenerater = (*JWTer)(nil)
	_ biz.TokenVerifyer  = (*JWTer)(nil)
)

var (
	ErrSignString           = errors.New("jwt sign string err")
	ErrParseJwtToken        = errors.New("parse jwt token failed")
	ErrInvalidToken         = errors.New("invalid token")
	InvalidUserID    uint64 = 0
)

// ProviderSet is token providers.
var ProviderSet = wire.NewSet(NewJWTer)

// JWTer 实现了redis.TokenProxy接口
type JWTer struct{}

func NewJWTer() *JWTer {
	return &JWTer{}
}

//使用jwt来生成token和验证token

func (j *JWTer) GenerateJwtToken(userID uint64, jwtSecret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = userID
	claims["timestamp"] = time.Now().Unix()
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("%w:%s", ErrSignString, err.Error())
	}
	return tokenString, nil
}

func (j *JWTer) VerifyJwtToken(tokenString, jwtSecret string) (uint64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return InvalidUserID, fmt.Errorf("%w:%s", ErrParseJwtToken, err.Error())
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return InvalidUserID, ErrInvalidToken

	}
	// 修改：断言为 float64 再转为 int32
	userIDFloat, ok := claims["userID"].(float64)
	if !ok {
		return InvalidUserID, ErrInvalidToken
	}

	userID := uint64(userIDFloat) // 转换为 uint64
	return userID, nil
}
