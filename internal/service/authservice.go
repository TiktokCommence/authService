package service

import (
	"context"
	pb "github.com/TiktokCommence/authService/api/auth/v1"
)

type AuthHandler interface {
	DeliverToken(ctx context.Context, userID uint64) (string, error)
	VerifyToken(ctx context.Context, tokenString *string, obj string, act string) (bool, error)
	GiveAuthority(ctx context.Context, userID uint64, role string) error
	RemoveAuthority(ctx context.Context, userID uint64, role string) error
}

type AuthServiceService struct {
	pb.UnimplementedAuthServiceServer
	authHandler AuthHandler
}

func NewAuthServiceService(authHandler AuthHandler) *AuthServiceService {
	return &AuthServiceService{
		authHandler: authHandler,
	}
}

func (s *AuthServiceService) DeliverTokenByRPC(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryResp, error) {
	token, err := s.authHandler.DeliverToken(ctx, req.GetUserId())
	if err != nil {
		return &pb.DeliveryResp{}, err
	}
	return &pb.DeliveryResp{
		Token: token,
	}, nil
}
func (s *AuthServiceService) VerifyTokenByRPC(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyResp, error) {

	obj, ok := TransformObj(req.GetObj())
	if !ok {
		return &pb.VerifyResp{
			Res: false,
		}, ErrParseObj
	}
	act, ok := TransformAct(req.GetAct())
	if !ok {
		return &pb.VerifyResp{
			Res: false,
		}, ErrParseAct
	}
	ok, err := s.authHandler.VerifyToken(ctx, req.Token, obj, act)
	if err != nil {
		return &pb.VerifyResp{
			Res: false,
		}, err
	}
	return &pb.VerifyResp{
		Res: ok,
	}, nil
}
func (s *AuthServiceService) AssignRole(ctx context.Context, req *pb.AssignRoleReq) (*pb.AssignResp, error) {
	role, ok := TransformRole(req.GetRole())
	if !ok {
		return &pb.AssignResp{
			Res: false,
		}, ErrParseRole
	}
	err := s.authHandler.GiveAuthority(ctx, req.GetUserId(), role)
	if err != nil {
		return &pb.AssignResp{
			Res: false,
		}, err
	}
	return &pb.AssignResp{
		Res: true,
	}, nil
}
func (s *AuthServiceService) RemoveRole(ctx context.Context, req *pb.RemoveRoleReq) (*pb.RemoveResp, error) {
	role, ok := TransformRole(req.GetRole())
	if !ok {
		return &pb.RemoveResp{
			Res: false,
		}, ErrParseRole
	}
	err := s.authHandler.RemoveAuthority(ctx, req.GetUserId(), role)
	if err != nil {
		return &pb.RemoveResp{
			Res: false,
		}, err
	}
	return &pb.RemoveResp{
		Res: true,
	}, nil
}
