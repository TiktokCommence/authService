package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/TiktokCommence/authService/internal/conf"
	token2 "github.com/TiktokCommence/authService/internal/token"
	"time"
)

var _ SignHandler = (*Signer)(nil)

//go:generate mockgen -destination=../mock/signer_mock.go -package=mock -source=./signer.go
var (
	ErrSetKey = errors.New("redis key set failed")
)

type TokenGenerater interface {
	GenerateJwtToken(userID uint64, jwtSecret string) (string, error)
}

type SetKeyer interface {
	SetKV(ctx context.Context, key string, value interface{}, expire time.Duration) error
}

type Signer struct {
	client SetKeyer
	t      TokenGenerater
	cf     *conf.Token
}

func NewSigner(client SetKeyer, t TokenGenerater, cf *conf.Token) *Signer {
	return &Signer{client: client, t: t, cf: cf}
}

func (s *Signer) SignToken(ctx context.Context, userID uint64) (string, error) {
	key := GenerateKey(userID)
	//redis中没有该用户的token,则创建
	jwtToken, err := s.t.GenerateJwtToken(userID, s.cf.Secret)

	// 如果签发token失败,重试3遍
	if errors.Is(err, token2.ErrSignString) {
		for i := 0; i < 2; i++ {
			jwtToken, err = s.t.GenerateJwtToken(userID, s.cf.Secret)
			if err == nil {
				break
			}
		}
	}
	//重试三遍后仍失败就返回错误
	if err != nil {
		return "", err
	}

	err = s.client.SetKV(ctx, key, jwtToken, s.getExpire())
	if err != nil {
		return "", fmt.Errorf("%w:%s", ErrSetKey, err.Error())
	}
	return jwtToken, nil
}

func (s *Signer) getExpire() time.Duration {
	return s.cf.Expiration.AsDuration()
}
