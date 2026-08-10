package main

import (
	"encoding/json"
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

type User_Response struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Email         string    `json:"email"`
	Token         string    `json:"token"`
	Refresh       string    `json:"refresh_token"`
	Is_chirpy_red bool      `json:"is_chirpy_red"`
}

type UserResponse struct {
	Email string `json:"email"`
}

func (cfg *ApiConfig) MakeUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Respond_err(w, 500, "error decoding request", err)
			return
		}

		pass, err := auth.HashPasswords(req.Password)
		if err != nil {
			Respond_err(w, 500, "error hasing given password", err)
			return
		}

		usr, err := cfg.dbQ.CreateUser(r.Context(), database.CreateUserParams{Email: req.Email, HashedPassword: pass})
		if err != nil {
			Respond_err(w, 500, "error inputting user into db", err)
			return
		}
		response := User_Response{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email, Is_chirpy_red: usr.IsChirpyRed}
		respondWithJSON(w, 201, response)
	})

}

func (cfg *ApiConfig) UserLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Respond_err(w, 500, "error decoding request", err)
			return
		}

		user, err := cfg.dbQ.GetUserViaEmail(r.Context(), req.Email)
		if err != nil {
			Respond_err(w, 500, "error fetching user via email", err)
			return
		}

		val, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
		if err != nil || !val {
			Respond_err(w, 500, "incorrect email or password", err)
			return
		}
		refresh := MakeRefreshToken()

		entry, err := cfg.dbQ.InsertToken(r.Context(), database.InsertTokenParams{Token: refresh, UserID: user.ID, ExpiresAt: time.Now().Add(1440 * time.Hour)})
		if err != nil {
			Respond_err(w, 500, "error inserting refresh token", err)
			return
		}

		tok, err := auth.MakeJwt(user.ID, cfg.JWT_SECRET, time.Duration(1)*time.Hour)
		if err != nil {
			Respond_err(w, 500, "error making jwt", err)
			return
		}

		response := User_Response{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, Token: tok, Refresh: entry.Token, Is_chirpy_red: user.IsChirpyRed}
		respondWithJSON(w, 200, response)
	})

}

func (cfg *ApiConfig) UpdateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r := Request{}
		if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
			Respond_err(w, 500, "error decoding request", err)
			return
		}
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Respond_err(w, 401, "error parsing header", err)
			return
		}

		user_id, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
		if err != nil {
			Respond_err(w, 401, "improper JWT", err)
			return
		}

		hashed, err := auth.HashPasswords(r.Password)
		if err != nil {
			Respond_err(w, 500, "error hashing password", err)
			return
		}
		updated, err := cfg.dbQ.UpdateUser(req.Context(), database.UpdateUserParams{Email: r.Email, HashedPassword: hashed, ID: user_id})
		if err != nil {
			Respond_err(w, 500, "error updating user into db", err)
			return
		}

		output := UserResponse{Email: updated.Email}
		respondWithJSON(w, 200, output)
	})

}
