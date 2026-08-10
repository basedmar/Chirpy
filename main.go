package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/basedmar/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error getting env vars")
	}
	db_url := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwt_secret := os.Getenv("JWT")
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		fmt.Println("error opening db handle")
	}
	defer db.Close()

	dbQuery := database.New(db)
	api := ApiConfig{filesrvHits: atomic.Int64{}, dbQ: dbQuery, Platform: platform, JWT_SECRET: jwt_secret}

	srvmux := http.NewServeMux()
	srvmux.Handle("POST /api/revoke", api.Revoke())
	srvmux.Handle("POST /api/refresh", api.Refreshtoken())
	srvmux.Handle("POST /api/login", api.UserLogin())
	srvmux.Handle("POST /api/users", api.MakeUser())
	srvmux.Handle("/app/", http.StripPrefix("/app/", api.MiddlewareHits(http.FileServer(http.Dir(".")))))
	srvmux.Handle("POST /admin/reset", api.ResetHits())
	srvmux.Handle("GET /admin/metrics", api.RetrieveHits())
	srvmux.Handle("GET /api/chirps", api.AllChirps())
	srvmux.Handle("POST /api/chirps", api.ChirpVal())
	srvmux.Handle("GET /api/chirps/{chirpID}", api.Getchirp())

	srvmux.HandleFunc("GET /api/healthz", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
	})

	Server := &http.Server{Handler: srvmux, Addr: ":8080"}
	Server.ListenAndServe()
}
