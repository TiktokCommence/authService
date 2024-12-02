package model

// subject
const (
	Traveler    = "traveler"   //游客，只拥有查看权利
	NormalUser  = "normalUser" //正常用户
	Merchant    = "merchant"   //商家
	Admin       = "admin"      //管理员
	BlackLister = "blackLister"

	ALLOW = "allow"
	DENY  = "deny"

	TravelerUserID = 0

	Create = "CREATE"
	Update = "UPDATE"
	Login  = "LOGIN"
	Logout = "LOGOUT"
	Delete = "DELETE"
	Get    = "GET"
	Clear  = "CLEAR"
	Pay    = "PAY"
	Cancel = "CANCEL"
	Settle = "SETTLE"
)

// object
var (
	U User
	I Item
	S ShoppingCart
	O Order
	P PayService
)
