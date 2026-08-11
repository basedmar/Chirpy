package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/basedmar/Chirpy/internal/auth"
)

func MakeRefreshToken() string {
	dat := make([]byte, 32)
	rand.Read(dat)
	value := hex.EncodeToString(dat)
	return value
}

func (cfg *ApiConfig) Refreshtoken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			Respond_err(w, 500, "error parsing token from header", err)
			return
		}

		res, err := cfg.dbQ.GetRefTokenUser(r.Context(), token)
		if err != nil {
			Respond_err(w, 401, "token doesnt exist or expired or revoked", err)
			return
		}
		if res.RevokedAt.Valid {
			Respond_err(w, 401, "token doesnt exist or expired or revoked", err)
			return
		}
		jwt, err := auth.MakeJwt(res.UserID, cfg.JWT_SECRET, time.Duration(1)*time.Hour)
		output := Refresh{Token: jwt}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		data, _ := json.Marshal(output)
		w.Write(data)
	})
}

func (cfg *ApiConfig) Revoke() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			Respond_err(w, 500, "error parsing token from header", err)
			return
		}
		err = cfg.dbQ.RevokeToken(r.Context(), token)
		if err != nil {
			Respond_err(w, 500, "error revoking refresh token", err)
			return
		}
		w.WriteHeader(204)
	})

}
