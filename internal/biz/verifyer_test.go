package biz

import (
	"context"
	"errors"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/TiktokCommence/authService/internal/mock"
	token2 "github.com/TiktokCommence/authService/internal/token"
	"google.golang.org/protobuf/types/known/durationpb"
	"sync"
	"testing"
	"time"
)

func TestVerifyer_checkTimeIfRich(t *testing.T) {
	type args struct {
		ttl    time.Duration
		expire time.Duration
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 time.Duration
	}{
		{"test1", args{ttl: time.Hour, expire: 4 * time.Hour}, false, 3 * time.Hour},
		{"test2", args{ttl: 40 * time.Minute, expire: 4 * time.Hour}, false, 3*time.Hour + 40*time.Minute},
		{"test3", args{ttl: 3*time.Hour + 10*time.Minute, expire: 4 * time.Hour}, true, 3*time.Hour + 10*time.Minute},
		{"test4", args{ttl: 2*time.Hour + 10*time.Minute, expire: 4 * time.Hour}, false, 3*time.Hour + 10*time.Minute},
	}
	v := &Verifyer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1 := v.checkTimeIfRich(tt.args.ttl, tt.args.expire)
			if got != tt.want {
				t.Errorf("checkTimeIfRich() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("checkTimeIfRich() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestVerifyer_getStoredToken(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mock.NewMockCacheProxy(ctrl)
	cache.EXPECT().GetValue(ctx, GenerateKey(111)).DoAndReturn(func(context.Context, string) (interface{}, error) {
		return "hello", nil
	})

	type args struct {
		ctx    context.Context
		userID int32
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{"test1", args{ctx: ctx, userID: 111}, "hello", nil},
	}

	v := &Verifyer{
		cache: cache,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.getStoredToken(tt.args.ctx, tt.args.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("getStoredToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getStoredToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyer_renewalToken(t *testing.T) {

	userID := []int32{111, 112, 113, 114, 115}
	token := []string{"hello1", "hello2", "hello3", "hello4", "hello5"}
	//userID := []int32{111}
	//token := []string{"hello1"}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mock.NewMockCacheProxy(ctrl)

	mockFuncs := []func(){
		func() {
			cache.EXPECT().ExistKey(ctx, GenerateKey(userID[0])).Return(true)
			cache.EXPECT().GetTTL(ctx, GenerateKey(userID[0])).Return(time.Hour, nil)
			cache.EXPECT().SetKV(ctx, GenerateKey(userID[0]), token[0], 3*time.Hour).Return(nil)
		},
		func() {
			cache.EXPECT().ExistKey(ctx, GenerateKey(userID[1])).Return(false)
			cache.EXPECT().SetKV(ctx, GenerateKey(userID[1]), token[1], 4*time.Hour).Return(nil)
		},
		func() {
			cache.EXPECT().ExistKey(ctx, GenerateKey(userID[2])).Return(true)
			cache.EXPECT().GetTTL(ctx, GenerateKey(userID[2])).Return(40*time.Minute, nil)
			cache.EXPECT().SetKV(ctx, GenerateKey(userID[2]), token[2], 3*time.Hour+40*time.Minute).Return(nil)
		},
		func() {
			cache.EXPECT().ExistKey(ctx, GenerateKey(userID[3])).Return(true)
			cache.EXPECT().GetTTL(ctx, GenerateKey(userID[3])).Return(2*time.Hour, nil)
			cache.EXPECT().SetKV(ctx, GenerateKey(userID[3]), token[3], 3*time.Hour).Return(nil)
		},
		func() {
			cache.EXPECT().ExistKey(ctx, GenerateKey(userID[4])).Return(true)
			cache.EXPECT().GetTTL(ctx, GenerateKey(userID[4])).Return(3*time.Hour+10*time.Second, nil)
		},
	}

	type args struct {
		ctx    context.Context
		userID int32
		token  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"test_1", args{ctx: ctx, userID: userID[0], token: token[0]}, nil},
		{"test_2", args{ctx: ctx, userID: userID[1], token: token[1]}, nil},
		{"test_3", args{ctx: ctx, userID: userID[2], token: token[2]}, nil},
		{"test_4", args{ctx: ctx, userID: userID[3], token: token[3]}, nil},
		{"test_5", args{ctx: ctx, userID: userID[4], token: token[4]}, nil},
	}

	dur := durationpb.New(4 * time.Hour)
	v := &Verifyer{
		cache: cache,
		cf:    &conf.Token{Expiration: dur},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncs[i]()
			if err := v.renewalToken(tt.args.ctx, tt.args.userID, tt.args.token); !errors.Is(err, tt.wantErr) {
				t.Errorf("renewalToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyer_VerifyToken(t *testing.T) {

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mock.NewMockCacheProxy(ctrl)
	tn := mock.NewMockTokenVerifyer(ctrl)

	type args struct {
		ctx         context.Context
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		wantID  int32
		wantErr error
	}{
		{"test_1", args{ctx: ctx, tokenString: "hello1"}, token2.InvalidUserID, token2.ErrInvalidToken},
		{"test_2", args{ctx: ctx, tokenString: "hello2"}, 112, ErrTokenNotFound},
		{"test_3", args{ctx: ctx, tokenString: "hello3"}, 113, ErrTokenInvalid},
		{"test_4", args{ctx: ctx, tokenString: "hello4"}, 114, nil},
	}

	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	mockFuncs := []func(){
		func() {
			tn.EXPECT().VerifyJwtToken("hello1", "hello").Return(token2.InvalidUserID, token2.ErrInvalidToken)
		},
		func() {
			tn.EXPECT().VerifyJwtToken("hello2", "hello").Return(int32(112), nil)
			cache.EXPECT().GetValue(ctx, GenerateKey(112)).Return("", redis.Nil)
		},
		func() {
			tn.EXPECT().VerifyJwtToken("hello3", "hello").Return(int32(113), nil)
			cache.EXPECT().GetValue(ctx, GenerateKey(113)).Return("hello2", nil)
		},
		func() {
			tn.EXPECT().VerifyJwtToken("hello4", "hello").Return(int32(114), nil)
			cache.EXPECT().GetValue(ctx, GenerateKey(114)).Return("hello4", nil)
			cache.EXPECT().ExistKey(ctx, GenerateKey(114)).DoAndReturn(func(context.Context, string) bool {
				defer wg.Done()
				return true
			})
			cache.EXPECT().GetTTL(ctx, GenerateKey(114)).DoAndReturn(func(ctx2 context.Context, string2 string) (time.Duration, error) {
				defer wg.Done()
				return 3*time.Hour + 10*time.Second, nil
			})
		},
	}

	v := &Verifyer{
		cache: cache,
		t:     tn,
		cf: &conf.Token{
			Secret:     "hello",
			Expiration: durationpb.New(4 * time.Hour),
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncs[i]()
			got0, err := v.VerifyToken(tt.args.ctx, tt.args.tokenString)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got0 != tt.wantID {
				t.Errorf("VerifyToken() got0 = %v, wantID %v", got0, tt.wantID)
			}
		})
	}
}
