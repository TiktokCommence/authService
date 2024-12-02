package casbin

import (
	"context"
	"errors"
	"fmt"
	"github.com/TiktokCommence/authService/internal/biz"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/TiktokCommence/authService/internal/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/wire"
)

var _ biz.RoleHandler = (*AuthCase)(nil)

var (
	ErrInvalidSubject = errors.New("invalid subject")
	ErrAssignRole     = errors.New("assign role failed")
	ErrSavePolicy     = errors.New("save policy failed")
	ErrCasBinEnforce  = errors.New("casbin enforce policy failed")
	ErrInvalidAct     = errors.New("invalid act")
	ErrRemoveRole     = errors.New("remove role failed")
)

// ProviderSet is casbin providers.
var ProviderSet = wire.NewSet(NewAuthCase)

type CheckActer interface {
	CheckActExist(act string) bool
}

type AuthCase struct {
	c *casbin.Enforcer
}

func NewAuthCase(cf *conf.Data) *AuthCase {
	// casbin使用gorm作为适配器
	// 使用 cf.Database.Driver 作为数据库,并且打开自动建表(若未存在)
	a, err := gormadapter.NewAdapter(cf.Casbin.Driver, cf.Casbin.Source, true)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer("authService/internal/casbin/model.conf", a)
	if err != nil {
		panic(err)
	}
	err = initPolicy(e)
	if err != nil {
		panic(err)
	}
	return &AuthCase{c: e}
}
func initPolicy(e *casbin.Enforcer) error {
	// 注意: 这个AddPolicies如果一旦遇到已经添加过的policy,它是直接返回added=true,不执行后面的policy,也就是说
	// 一旦你在数据库中删去了老的规则,在这里添加了新的规则,实际上是不会执行的,如果是加在开头,可能有用
	_, err := e.AddPolicies(
		[][]string{
			// *代表所有
			{model.Admin, "*", "*", model.ALLOW},

			{model.BlackLister, "*", "*", model.DENY},

			//{model.Traveler,model.U.String(),model.U.Create(),model.ALLOW},
			//{model.Traveler,model.U.String(),model.U.Update(),model.DENY},
			//{model.Traveler,model.U.String(),model.U.Login(),model.ALLOW},
			//{model.Traveler,model.U.String(),model.U.Logout(),model.DENY},
			//{model.Traveler,model.U.String(),model.U.Delete(),model.DENY},
			//{model.Traveler,model.U.String(),model.U.GetInfo(),model.DENY},

			//由于在model.conf中
			//[policy_effect]
			//e = some(where (p.eft == allow))
			//由于游客大部分的操作都是不被允许的，为了简洁起见，所以就先否定所有，在从中选出其中可以执行的操作
			{model.Traveler, "*", "*", model.DENY},
			{model.Traveler, model.U.String(), model.U.Create(), model.ALLOW},
			{model.Traveler, model.U.String(), model.U.Login(), model.ALLOW},
			{model.Traveler, model.I.String(), model.I.GetInfo(), model.ALLOW},

			{model.NormalUser, model.U.String(), model.U.Create(), model.ALLOW},
			{model.NormalUser, model.U.String(), model.U.Update(), model.ALLOW},
			{model.NormalUser, model.U.String(), model.U.Login(), model.ALLOW},
			{model.NormalUser, model.U.String(), model.U.Logout(), model.ALLOW},
			{model.NormalUser, model.U.String(), model.U.Delete(), model.ALLOW},
			{model.NormalUser, model.U.String(), model.U.GetInfo(), model.ALLOW},
			{model.NormalUser, model.I.String(), model.I.Create(), model.DENY},
			{model.NormalUser, model.I.String(), model.I.GetInfo(), model.ALLOW},
			{model.NormalUser, model.I.String(), model.I.UpdateInfo(), model.DENY},
			{model.NormalUser, model.I.String(), model.I.Delete(), model.DENY},
			{model.NormalUser, model.S.String(), "*", model.ALLOW},
			{model.NormalUser, model.O.String(), "*", model.ALLOW},
			{model.NormalUser, model.P.String(), "*", model.ALLOW},

			{model.Merchant, "*", "*", model.DENY},
			{model.Merchant, model.U.String(), "*", model.ALLOW},
			{model.Merchant, model.I.String(), "*", model.ALLOW},
		})
	if err != nil {
		return err
	}

	// 默认添加0为traveler
	_, err = e.AddGroupingPolicy(fmt.Sprintf("%d", model.TravelerUserID), model.Traveler)
	if err != nil {
		return fmt.Errorf("%w:%s", ErrAssignRole, err.Error())
	}

	err = e.SavePolicy()
	if err != nil {
		return err
	}
	return nil
}

// AssignAuthority 为用户分配权限
// 成功返回true,失败返回false
func (a *AuthCase) AssignAuthority(ctx context.Context, userID uint64, role string) error {
	err := a.checkRole(role)
	if err != nil {
		return err
	}
	id := a.generateUserID(userID)
	// 为用户分配角色
	_, err = a.c.AddGroupingPolicy(id, role)
	if err != nil {
		return fmt.Errorf("%w:%s", ErrAssignRole, err.Error())
	}
	//保存
	if err := a.c.SavePolicy(); err != nil {
		return fmt.Errorf("%w:%s", ErrSavePolicy, err.Error())
	}
	return nil
}

// VerifyAuthority 验证 userID对于obj资源是否可以执行act操作
// 如果执行成功,返回验证结果
// 如果操作不合法或者obj不合法,就会返回error("invalid act")
// 如果执行失败就返回error("casbin enforce policy failed")
// obj及对应的act在下面
// user : "CREATE","LOGIN","LOGOUT","DELETE","UPDATE","GET"
// shopping_cart : "CREATE","CLEAR","GET"
// pay : "PAY","CANCEL"
// order : "CREATE","UPDATE","SETTLE"
// item : "CREATE","DELETE","UPDATE","GET"
func (a *AuthCase) VerifyAuthority(ctx context.Context, userID uint64, obj, act string) (bool, error) {
	if !a.checkAct(obj, act) {
		return false, ErrInvalidAct
	}

	id := a.generateUserID(userID)
	ok, err := a.c.Enforce(id, obj, act)
	if err != nil {
		return false, fmt.Errorf("%w:%s", ErrCasBinEnforce, err.Error())
	}
	return ok, nil
}

func (a *AuthCase) RemoveAuthority(ctx context.Context, userID uint64, role string) error {
	err := a.checkRole(role)
	if err != nil {
		return err
	}
	id := a.generateUserID(userID)
	_, err = a.c.RemoveGroupingPolicy(id, role)
	if err != nil {
		return fmt.Errorf("%w:%s", ErrRemoveRole, err.Error())
	}
	return nil
}

// 判断角色是否合法
func (a *AuthCase) checkRole(role string) error {
	switch role {
	case model.NormalUser, model.Merchant, model.BlackLister:
	default:
		return ErrInvalidSubject
	}
	return nil
}

// 检查对obj是否有act操作
// 使用映射来减少 switch 语句
func (a *AuthCase) checkAct(obj, act string) bool {
	// 定义一个映射，直接将 obj 映射到具体的类型
	acters := map[string]CheckActer{
		model.U.String(): model.User{},
		model.I.String(): model.Item{},
		model.O.String(): model.Order{},
		model.P.String(): model.PayService{},
		model.S.String(): model.ShoppingCart{},
	}

	// 获取对应类型的实例
	checker, exists := acters[obj]
	if !exists {
		return false
	}

	// 调用 CheckActExist 方法
	return checker.CheckActExist(act)
}

func (a *AuthCase) generateUserID(userID uint64) string {
	return fmt.Sprintf("%d", userID)
}
