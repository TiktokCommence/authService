package biz

import (
	"context"
	"errors"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/go-redis/redis"
	"time"
)

var _ VerifyHandler = (*Verifyer)(nil)

//go:generate mockgen -destination=../mock/verifyer_mock.go -package=mock -source=./verifyer.go
var (
	ErrTokenNotFound = errors.New("don't find the token")
	ErrTokenInvalid  = errors.New("the token is invalid")
)

type TokenVerifyer interface {
	VerifyJwtToken(tokenString, jwtSecret string) (uint64, error)
}

type CacheProxy interface {
	SetKV(ctx context.Context, key string, value interface{}, expire time.Duration) error
	GetValue(ctx context.Context, key string) (interface{}, error)
	ExistKey(ctx context.Context, key string) bool
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

type Verifyer struct {
	cache CacheProxy
	t     TokenVerifyer
	cf    *conf.Token
}

func NewVerifier(client CacheProxy, t TokenVerifyer, cf *conf.Token) *Verifyer {
	return &Verifyer{cache: client, t: t, cf: cf}
}

// VerifyToken 检查token是否合法
// 如果不合法就返回false
// 如果合法,异步为token续期
func (v *Verifyer) VerifyToken(ctx context.Context, tokenString string) (uint64, error) {
	userID, err := v.t.VerifyJwtToken(tokenString, v.cf.Secret)
	if err != nil {
		return userID, err
	}

	storedToken, err := v.getStoredToken(ctx, userID)
	// 如果该token未找到
	if errors.Is(err, redis.Nil) {
		return userID, ErrTokenNotFound
	}
	// 如果传入的token与查到的token不一致也报错
	if storedToken != tokenString {
		return userID, ErrTokenInvalid
	}

	// 异步去为token续期,根据剩余时间的多少
	go func() {
		// 添加重试机制
		for i := 0; i < 2; i++ {
			err := v.renewalToken(ctx, userID, storedToken)
			if err == nil {
				break
			}
		}
	}()
	return userID, nil
}

// 为token续期逻辑
func (v *Verifyer) renewalToken(ctx context.Context, userID uint64, token string) error {
	key := GenerateKey(userID)

	// verify的时候存在,但renewal的时候不存在,说明原本就是快要过期的状态,这时应该直接续期
	if !v.cache.ExistKey(ctx, key) {
		err := v.cache.SetKV(ctx, key, token, v.getExpire())
		return err
	}

	//获取userID的剩余过期时间
	ttl, err := v.cache.GetTTL(ctx, key)
	if err != nil {
		return err
	}

	// 检验剩余时间的程度
	// 根据不同的程度,来增加剩余过期时间
	rich, newTTL := v.checkTimeIfRich(ttl, v.getExpire())
	if !rich {
		err = v.cache.SetKV(ctx, key, token, newTTL)
		return err
	}
	return nil
}

// 检查剩余过期时间是否充裕
// 如果充裕返回true
// 不充裕返回false和应该设置的过期时间
func (v *Verifyer) checkTimeIfRich(ttl time.Duration, expire time.Duration) (bool, time.Duration) {
	if ttl < expire/4 {
		return false, ttl + 3*expire/4
	} else if ttl < expire/2 {
		return false, ttl + expire/2
	} else if ttl < 3*expire/4 {
		return false, ttl + expire/4
	}
	return true, ttl
}

func (v *Verifyer) getExpire() time.Duration {
	return v.cf.Expiration.AsDuration()
}

func (v *Verifyer) getStoredToken(ctx context.Context, userID uint64) (string, error) {
	key := GenerateKey(userID)
	res, err := v.cache.GetValue(ctx, key)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}
