package service

import (
	"errors"
	pb "github.com/TiktokCommence/authService/api/auth/v1"
	"github.com/TiktokCommence/authService/internal/model"
	"github.com/google/wire"
)

var (
	// ProviderSet is service providers.
	ProviderSet = wire.NewSet(NewAuthServiceService)
	ObjMap      = map[pb.Obj]string{
		pb.Obj_User:          model.U.String(),
		pb.Obj_Item:          model.I.String(),
		pb.Obj_Order:         model.O.String(),
		pb.Obj_PayService:    model.P.String(),
		pb.Obj_Shopping_cart: model.S.String(),
	}
	ActMap = map[pb.Act]string{
		pb.Act_Create: model.Create,
		pb.Act_Login:  model.Login,
		pb.Act_Logout: model.Logout,
		pb.Act_Update: model.Update,
		pb.Act_Delete: model.Delete,
		pb.Act_Get:    model.Get,
		pb.Act_Clear:  model.Clear,
		pb.Act_Pay:    model.Pay,
		pb.Act_Cancel: model.Cancel,
		pb.Act_Settle: model.Settle,
	}
	RoleMap = map[pb.Role]string{
		pb.Role_BlackLister: model.BlackLister,
		pb.Role_NormalUser:  model.NormalUser,
		pb.Role_Merchant:    model.Merchant,
	}
	ErrParseObj  = errors.New("invalid object")
	ErrParseAct  = errors.New("invalid action")
	ErrParseRole = errors.New("invalid role")
)

func TransformObj(o pb.Obj) (string, bool) {
	obj, ok := ObjMap[o]
	if !ok {
		return "", false
	}
	return obj, true
}

func TransformAct(a pb.Act) (string, bool) {
	act, ok := ActMap[a]
	if !ok {
		return "", false
	}
	return act, true
}

func TransformRole(a pb.Role) (string, bool) {
	role, ok := RoleMap[a]
	if !ok {
		return "", false
	}
	return role, true
}
