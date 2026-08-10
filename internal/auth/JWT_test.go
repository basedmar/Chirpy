package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	id := uuid.New()
	jwt, err := MakeJwt(id, "aoseuihfaeuoifhae", 20*time.Minute)
	if err != nil {
		t.Errorf("error making jwt %s", err)
	}
	val, err := ValidateJWT(jwt, "aoseuihfaeuoifhae")
	if val != id {
		t.Errorf("two user ids from jwts dont match !!")
	}

}

func TestBad(t *testing.T) {
	id := uuid.New()
	jwt, err := MakeJwt(id, "aoseuihfaeuoifhae", 1*time.Millisecond)
	time.Sleep(1 * time.Second)
	if err != nil {
		t.Errorf("error making jwt %s", err)
	}
	_, err = ValidateJWT(jwt, "aoseuihfaeuoifhae")
	if err != nil {
		valer := fmt.Sprintf("%s", err)
		if valer == "token has invalid claims: token is expired" {
			t.Logf("good token good bad")
		}
	}
}

func TestBadsecret(t *testing.T) {
	id := uuid.New()
	jwt, err := MakeJwt(id, "aoseuihfaeuoifhae", 50*time.Minute)
	if err != nil {
		t.Errorf("error making jwt %s", err)
	}
	_, err = ValidateJWT(jwt, "aosrgsgrgsrsgrgrsgsrgrgrsgrsuihfaeuoifhae")
	if err != nil {
		errorr := fmt.Sprintf("%s", err)
		if errorr == "token signature is invalid: signature is invalid" {
			t.Logf("good token good bad")
		}

	}
}
