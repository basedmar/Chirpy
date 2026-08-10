package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")

	if token == "" {
		return "", fmt.Errorf("header doesnt exist for jwt")
	}
	if len(token) > 7 {
		split := strings.Split(token, " ")
		tokenn := split[1]

		return tokenn, nil
	} else {
		return "", fmt.Errorf("string not long enough / something wrong with formatting")
	}
}

func HashPasswords(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	result, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	if result {
		return true, nil
	}
	return false, nil
}

func MakeJwt(userid uuid.UUID, tokensecret string, expiresin time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresin)),
		Subject:   userid.String(),
	}
	key := []byte(tokensecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenn, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenn, nil
}

func ValidateJWT(tokenstring, tokensecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenstring, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokensecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	id := claims.Subject
	made, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, err
	}
	return made, nil
}
