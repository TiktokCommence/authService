package model

//pay服务的相关操作
//只是用来返回其操作的name

// PayService 实现了casbin.CheckActer 接口
type PayService struct{}

func (p PayService) String() string {
	return "pay"
}

// act

func (p PayService) Pay() string {
	return Pay
}
func (p PayService) Cancel() string {
	return Cancel
}

// CheckActExist 检查对于payService,act是否合法
func (p PayService) CheckActExist(act string) bool {
	switch act {
	case p.Pay(), p.Cancel():
	default:
		return false
	}
	return true
}
