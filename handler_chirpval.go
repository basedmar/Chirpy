package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/basedmar/Chirpy/internal/auth"
	"github.com/basedmar/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpReq struct {
	Body string `json:"body"`
}

type chirpSend struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func filterbadword(input string) string {
	filter := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	split := strings.Split(input, " ")
	var clean []string
	for _, v := range split {
		lower := strings.ToLower(v)
		if _, ok := filter[lower]; ok {
			clean = append(clean, "****")
			continue
		}
		clean = append(clean, v)
	}
	return strings.Join(clean, " ")
}

func (cfg *ApiConfig) ChirpVal() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		chirps := chirpReq{}
		if err := decoder.Decode(&chirps); err != nil {
			Respond_err(w, 500, "error decoding request body", err)
			return
		}

		if len(chirps.Body) > 140 {
			Respond_err(w, 500, "Chirp too long!", nil)
			return
		}

		clean_body := filterbadword(chirps.Body)

		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Respond_err(w, 500, "error paring header", err)
		}
		id, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
		if err != nil {
			Respond_err(w, 401, "BAD JWT!", err)
			return
		}
		request, err := cfg.dbQ.CreateChirp(req.Context(), database.CreateChirpParams{Body: clean_body, UserID: id})
		if err != nil {
			Respond_err(w, 500, "error creating chirp", err)
		}
		sender := chirpSend{ID: request.ID, CreatedAt: request.CreatedAt, UpdatedAt: request.UpdatedAt, Body: request.Body, UserID: request.UserID}
		respondWithJSON(w, 201, sender)
	})
}

func (cfg *ApiConfig) AllChirps() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s := req.URL.Query().Get("author_id")
		b := req.URL.Query().Get("sort")
		chirps, err := cfg.dbQ.GetAllChirps(req.Context())
		if err != nil {
			Respond_err(w, 500, "error fetching chirps from db", err)
			return
		}
		var id uuid.UUID
		if s != "" {
			id, err = uuid.Parse(s)
			if err != nil {
				Respond_err(w, 401, "invalid id for user", err)
				return
			}
		}
		var final []chirpSend
		if s != "" {
			for _, v := range chirps {
				if v.UserID == id {
					final = append(final, chirpSend{ID: v.UserID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Body: v.Body, UserID: v.UserID})
				} else {
					continue
				}
			}
		} else {
			for _, v := range chirps {
				final = append(final, chirpSend{ID: v.UserID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Body: v.Body, UserID: v.UserID})
			}
		}

		if b == "asc" {
			sort.Slice(final, func(i, j int) bool {
				output := final[i].CreatedAt.Compare(final[j].CreatedAt)
				return output == -1
			})
		} else if b == "desc" {
			sort.Slice(final, func(i, j int) bool {
				output := final[i].CreatedAt.Compare(final[j].CreatedAt)
				return output == 1
			})
		}
		send, err := json.Marshal(final)
		if err != nil {
			Respond_err(w, 500, "error marshaling chirps from fetchallchirps", err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(send)
	})

}

func (cfg *ApiConfig) Getchirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			fmt.Println("couldnt parse id")
		}
		chirp, err := cfg.dbQ.GetChirp(req.Context(), id)
		if err != nil {
			fmt.Println("error getting chirp")
			Respond_err(w, 404, "coudlnt find chirp", err)
		}
		chirper := chirpSend{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
		respondWithJSON(w, 200, chirper)
	})

}

func (cfg *ApiConfig) Deletechirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			Respond_err(w, 500, "error parsing path for id", err)
			return
		}
		chirp, err := cfg.dbQ.GetChirp(req.Context(), id)
		if err != nil {
			Respond_err(w, 404, "coudlnt find chirp", err)
			return
		}
		jwt, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Respond_err(w, 401, "error parsing token from header", err)
			return
		}
		val, err := auth.ValidateJWT(jwt, cfg.JWT_SECRET)
		if err != nil {
			Respond_err(w, 401, "invalid jwt", err)
			return
		}
		if chirp.UserID != val {
			Respond_err(w, 403, "jwt user does not own specified chirp for deletion", err)
			return
		}
		err = cfg.dbQ.DeleteChirp(req.Context(), chirp.ID)
		if err != nil {
			Respond_err(w, 500, "error deleting chirp", err)
			return
		}
		w.WriteHeader(204)
	})

}
