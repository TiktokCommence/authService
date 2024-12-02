package model

//order服务的相关操作
//只是用来返回其操作的name

// Order 实现了casbin.CheckActer 接口
type Order struct{}

func (o Order) String() string {
	return "order"
}

// act

func (o Order) Create() string {
	return Create
}
func (o Order) Update() string {
	return Update
}

// Settle 结算订单
func (o Order) Settle() string {
	return Settle
}

// CheckActExist 检查对于order,act是否合法
func (o Order) CheckActExist(act string) bool {
	switch act {
	case o.Create(), o.Update(), o.Settle():
	default:
		return false
	}
	return true
}
