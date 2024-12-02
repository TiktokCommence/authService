package token

import (
	"fmt"
	"testing"
)

func TestJwt(t *testing.T) {
	j := NewJWTer()
	var (
		UserID uint64 = 111
		Secret        = "happy"
	)
	tokenString, err := j.GenerateJwtToken(UserID, Secret)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokenString)
	res, err := j.VerifyJwtToken(tokenString, Secret)
	if err != nil {
		t.Error(err)
	}
	if res != UserID {
		t.Error("verify failed")
	}
}
