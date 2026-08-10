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

func GetApiKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("header doesnt exist for polka key check")
	}
	split := strings.Split(header, " ")
	if len(split) != 2 {
		return "", fmt.Errorf("inproper auth header")
	}

	token := split[1]
	return token, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")

	if header == "" {
		return "", fmt.Errorf("header doesnt exist for jwt")
	}

	split := strings.Split(header, " ")
	if len(split) != 2 {
		return "", fmt.Errorf("inproper auth header")
	}

	token := split[1]
	return token, nil

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
