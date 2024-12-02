package model

//shopping_cart服务的相关操作
//只是用来返回其操作的name

// ShoppingCart 实现了casbin.CheckActer 接口
type ShoppingCart struct{}

func (s ShoppingCart) String() string {
	return "shopping_cart"
}

// act

func (s ShoppingCart) Create() string {
	return Create
}
func (s ShoppingCart) Clear() string {
	return Clear
}
func (s ShoppingCart) GetInfo() string {
	return Get
}

// CheckActExist 检查对于ShoppingCart,act是否合法
func (s ShoppingCart) CheckActExist(act string) bool {
	switch act {
	case s.Create(), s.Clear(), s.GetInfo():
	default:
		return false
	}
	return true
}
