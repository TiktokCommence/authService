package biz

import (
	"context"
	"fmt"
	"github.com/TiktokCommence/authService/internal/model"
	"github.com/TiktokCommence/authService/internal/service"
	"github.com/google/wire"
)

var _ service.AuthHandler = (*AuthUserCase)(nil)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAuthUserCase, NewSigner, NewVerifier)

func GenerateKey(userID uint64) string {
	return fmt.Sprintf("auth:%d", userID)
}

type RoleHandler interface {
	AssignAuthority(ctx context.Context, userID uint64, role string) error
	VerifyAuthority(ctx context.Context, userID uint64, obj, act string) (bool, error)
	RemoveAuthority(ctx context.Context, userID uint64, role string) error
}

type SignHandler interface {
	SignToken(ctx context.Context, userID uint64) (string, error)
}

type VerifyHandler interface {
	VerifyToken(ctx context.Context, tokenString string) (uint64, error)
}

type AuthUserCase struct {
	r RoleHandler
	s SignHandler
	v VerifyHandler
}

func NewAuthUserCase(r RoleHandler, s SignHandler, v VerifyHandler) *AuthUserCase {
	return &AuthUserCase{r: r, s: s, v: v}
}
func (a *AuthUserCase) DeliverToken(ctx context.Context, userID uint64) (string, error) {
	return a.s.SignToken(ctx, userID)
}
func (a *AuthUserCase) VerifyToken(ctx context.Context, tokenString *string, obj string, act string) (bool, error) {
	var (
		userID uint64
		err    error
	)

	if tokenString == nil {
		userID = model.TravelerUserID
	} else {
		userID, err = a.v.VerifyToken(ctx, *tokenString)
		if err != nil {
			return false, err
		}
	}

	ok, err := a.r.VerifyAuthority(ctx, userID, obj, act)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (a *AuthUserCase) GiveAuthority(ctx context.Context, userID uint64, role string) error {
	return a.r.AssignAuthority(ctx, userID, role)
}

func (a *AuthUserCase) RemoveAuthority(ctx context.Context, userID uint64, role string) error {
	return a.r.RemoveAuthority(ctx, userID, role)
}
