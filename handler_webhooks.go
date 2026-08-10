package main

import (
	"encoding/json"
	"net/http"

	"github.com/basedmar/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type Webhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *ApiConfig) UpgradeUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body := Webhook{}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			Respond_err(w, 500, "error decoding req body", err)
			return
		}
		if body.Event != "user.upgraded" {
			Respond_err(w, 204, "not upgrade event", nil)
			return
		}
		key, err := auth.GetApiKey(req.Header)
		if err != nil {
			Respond_err(w, 401, "bad formatting for api key", err)
			return
		}
		if key != cfg.Polka_key {
			Respond_err(w, 401, "incorrect api key", err)
			return
		}

		id, err := uuid.Parse(body.Data.UserID)
		if err != nil {
			Respond_err(w, 401, "given id is not in proper uuid format", err)
			return
		}
		if body.Event == "user.upgraded" {
			err = cfg.dbQ.UpgradeUser(req.Context(), id)
			if err != nil {
				Respond_err(w, 401, "user could not be deleted / user is not in the db", err)
			}
		}
		w.WriteHeader(204)
		empty_body, err := json.Marshal(struct{}{})
		if err != nil {
			Respond_err(w, 500, "error making body", err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(empty_body)
	})

}
