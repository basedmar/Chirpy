package main

import (
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) MiddlewareHits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.filesrvHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) RetrieveHits() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		val := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.filesrvHits.Load())
		w.Write([]byte(val))
	})
}

func (cfg *ApiConfig) ResetHits() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Platform != "DEV" {
			Respond_err(w, 403, "not dev", nil)
			return
		}
		cfg.dbQ.DeleteUsers(r.Context())
		cfg.filesrvHits.Store(0)
	})
}
