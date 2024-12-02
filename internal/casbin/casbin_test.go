package casbin

import (
	"context"
	"errors"
	"github.com/TiktokCommence/authService/internal/conf"
	"testing"
)

func TestAuthCase_AssignAuthority(t *testing.T) {

	type args struct {
		userID int32
		role   string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"test1", args{userID: 113, role: "normalUser"}, nil},
		{"test2", args{userID: 114, role: "merchant"}, nil},
		{"test3", args{userID: 115, role: "blackLister"}, nil},
		{"test4", args{userID: 116, role: "hahaha"}, ErrInvalidSubject},
	}
	a := test_init()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.AssignAuthority(context.Background(), tt.args.userID, tt.args.role); got != tt.want {
				t.Errorf("AssignAuthority() = %v, want %v", got, tt.want)
			}
		})
	}
}
func test_init() *AuthCase {
	cc := &conf.Data_CasbinConf{Driver: "mysql", Source: "root:12345678@tcp(127.0.0.1:13306)/casbin"}

	return NewAuthCase(&conf.Data{Casbin: cc})
}

func Test_checkAct(t *testing.T) {
	type args struct {
		obj string
		act string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test_user_1", args{"user", "CREATE"}, true},
		{"test_user_2", args{"user", "LOGIN"}, true},
		{"test_user_3", args{"user", "LOGOUT"}, true},
		{"test_user_4", args{"user", "LOGOUT"}, true},
		{"test_user_5", args{"user", "DELETE"}, true},
		{"test_user_6", args{"user", "UPDATE"}, true},
		{"test_user_7", args{"user", "GET"}, true},
		{"test_user_8", args{"user", "hello"}, false},

		{"test_shoppingcart_1", args{"shopping_cart", "CREATE"}, true},
		{"test_shoppingcart_2", args{"shopping_cart", "GET"}, true},
		{"test_shoppingcart_3", args{"shopping_cart", "CLEAR"}, true},
		{"test_shoppingcart_4", args{"shopping_cart", "hello"}, false},

		{"test_pay_1", args{"pay", "PAY"}, true},
		{"test_pay_2", args{"pay", "CANCEL"}, true},
		{"test_pay_3", args{"pay", "hello"}, false},

		{"test_order_1", args{"order", "CREATE"}, true},
		{"test_order_2", args{"order", "UPDATE"}, true},
		{"test_order_3", args{"order", "SETTLE"}, true},
		{"test_order_4", args{"order", "hello"}, false},

		{"test_item_1", args{"item", "CREATE"}, true},
		{"test_item_2", args{"item", "UPDATE"}, true},
		{"test_item_3", args{"item", "DELETE"}, true},
		{"test_item_4", args{"item", "GET"}, true},
		{"test_item_5", args{"item", "hello"}, false},

		{"test_unknown_name_1", args{"hahah", "hello"}, false},
	}
	var a = new(AuthCase)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.checkAct(tt.args.obj, tt.args.act); got != tt.want {
				t.Errorf("checkAct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthCase_VerifyAuthority(t *testing.T) {
	type args struct {
		userID int32
		obj    string
		act    string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr error
	}{
		// userID=111 is admin
		{"test_admin", args{111, "user", "CREATE"}, true, nil},
		// userID=115 is blackLister
		{"test_blacklisted", args{115, "user", "CREATE"}, false, nil},
		// userID=112 is traveler
		{"test_traveler_1", args{112, "user", "CREATE"}, true, nil},
		{"test_traveler_2", args{112, "user", "LOGIN"}, true, nil},
		{"test_traveler_3", args{112, "item", "GET"}, true, nil},
		{"test_traveler_4", args{112, "user", "DELETE"}, false, nil},

		// userID=113 is normalUser
		{"test_normal_user_1", args{113, "user", "CREATE"}, true, nil},
		{"test_normal_user_2", args{113, "user", "DELETE"}, true, nil},
		{"test_normal_user_3", args{113, "user", "GET"}, true, nil},
		{"test_normal_user_4", args{113, "user", "LOGIN"}, true, nil},
		{"test_normal_user_5", args{113, "user", "LOGOUT"}, true, nil},
		{"test_normal_user_6", args{113, "user", "UPDATE"}, true, nil},
		{"test_normal_user_7", args{113, "pay", "PAY"}, true, nil},
		{"test_normal_user_8", args{113, "pay", "CANCEL"}, true, nil},
		{"test_normal_user_9", args{113, "item", "GET"}, true, nil},
		{"test_normal_user_10", args{113, "item", "CREATE"}, false, nil},

		// userID=114 is merchant
		{"test_merchant_1", args{114, "user", "CREATE"}, true, nil},
		{"test_merchant_2", args{114, "user", "DELETE"}, true, nil},
		{"test_merchant_3", args{114, "user", "GET"}, true, nil},
		{"test_merchant_4", args{114, "user", "LOGIN"}, true, nil},
		{"test_merchant_5", args{114, "user", "LOGOUT"}, true, nil},
		{"test_merchant_6", args{114, "user", "UPDATE"}, true, nil},
		{"test_merchant_7", args{114, "pay", "PAY"}, false, nil},
		{"test_merchant_8", args{114, "pay", "CANCEL"}, false, nil},
		{"test_merchant_9", args{114, "item", "CREATE"}, true, nil},
		{"test_merchant_10", args{114, "item", "GET"}, true, nil},

		// error situation
		{"test_error_1", args{114, "hahah", "get"}, false, ErrInvalidAct},
		{"test_error_2", args{114, "user", "CLEAR"}, false, ErrInvalidAct},
	}
	a := test_init()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.VerifyAuthority(context.Background(), tt.args.userID, tt.args.obj, tt.args.act)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("VerifyAuthority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyAuthority() got = %v, want %v", got, tt.want)
			}
		})
	}
}
