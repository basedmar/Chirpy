package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/basedmar/Chirpy/internal/auth"
	"github.com/basedmar/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Request struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Refresh struct {
	Token string `json:"token"`
}

type Userr struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	Refresh   string    `json:"refresh_token"`
}

func (cfg *ApiConfig) MakeUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Print("error decoding request!")
			w.WriteHeader(500)
			return
		}
		pass, err := auth.HashPasswords(req.Password)
		if err != nil {
			log.Printf("error hashing password %s", err)
			w.WriteHeader(500)
			return
		}

		usr, err := cfg.dbQ.CreateUser(r.Context(), database.CreateUserParams{Email: req.Email, HashedPassword: pass})
		if err != nil {
			log.Printf("error making user in db! %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(201)
		realer := Userr{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email}
		data, err := json.Marshal(realer)
		w.Write(data)
	})

}

func (cfg *ApiConfig) UserLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := Request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Print("error decoding request!")
			w.WriteHeader(500)
			return
		}
		user, err := cfg.dbQ.GetUserViaEmail(r.Context(), req.Email)
		if err != nil {
			log.Print("error fetching user from db via email!")
			w.WriteHeader(500)
			return
		}
		refresh := MakeRefreshToken()

		entry, err := cfg.dbQ.InsertToken(r.Context(), database.InsertTokenParams{Token: refresh, UserID: user.ID, ExpiresAt: time.Now().Add(1440 * time.Hour)})
		val, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
		if val {

			tok, err := auth.MakeJwt(user.ID, cfg.JWT_SECRET, time.Duration(1)*time.Hour)
			if err != nil {
				fmt.Println("error making jwt token")
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)
			realer := Userr{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, Token: tok, Refresh: entry.Token}
			data, _ := json.Marshal(realer)
			w.Write(data)
		} else {
			Respond_err(w, 401, "Incorrect email or password", nil)
			return
		}
	})

}

func (cfg *ApiConfig) Refreshtoken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			fmt.Println("error getting refresh token from header")
		}

		res, err := cfg.dbQ.GetRefTokenUser(r.Context(), token)
		if err != nil {
			Respond_err(w, 401, "token doesnt exist or expired or revoked", err)
		}
		if res.RevokedAt.Valid {
			Respond_err(w, 401, "token doesnt exist or expired or revoked", err)
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
			fmt.Println("error getting refresh token from header")
		}
		err = cfg.dbQ.RevokeToken(r.Context(), token)
		if err != nil {
			fmt.Println("error revoking")
		}
		w.WriteHeader(204)
	})

}
