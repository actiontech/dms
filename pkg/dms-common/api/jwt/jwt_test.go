package jwt

import (
	"testing"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

	"github.com/golang-jwt/jwt/v4"
)

func TestGenJwtToken(t *testing.T) {
	token, err := GenJwtToken(WithUserId("999999"))
	if err != nil {
		t.Errorf("failed to sign the token: %v", err)
	}
	tokenAfterParse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return dmsCommonV1.JwtSigningKey, nil
	})
	if err != nil {
		t.Errorf("failed to parse the token: %v", err)
	}

	uid, err := ParseUserFromToken(tokenAfterParse)
	if uid != 999999 || err != nil {
		t.Errorf("failed to parse the token: %v", err)
	}
}
