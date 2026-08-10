package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/basedmar/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type ApiConfig struct {
	filesrvHits atomic.Int64
	dbQ         *database.Queries
	Platform    string
	JWT_SECRET  string
	Polka_key   string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error getting env vars")
	}
	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		log.Fatal("error loading db url")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("error loading platform")
	}
	jwt_secret := os.Getenv("JWT")
	if jwt_secret == "" {
		log.Fatal("error loading jwt secret phrase")
	}
	polka_key := os.Getenv("POLKA_KEY")
	if polka_key == "" {
		log.Fatal("error loading polka webhoook api key")
	}

	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatalf("error opening handle to db: %s", err)
	}
	defer db.Close()

	dbQuery := database.New(db)
	api := ApiConfig{filesrvHits: atomic.Int64{}, dbQ: dbQuery, Platform: platform, JWT_SECRET: jwt_secret, Polka_key: polka_key}

	srvmux := http.NewServeMux()

	srvmux.Handle("POST /api/users", api.MakeUser())
	srvmux.Handle("PUT /api/users", api.UpdateUser())
	srvmux.Handle("POST /api/login", api.UserLogin())

	srvmux.Handle("DELETE /api/chirps/{chirpID}", api.Deletechirp())
	srvmux.Handle("GET /api/chirps/{chirpID}", api.Getchirp())
	srvmux.Handle("GET /api/chirps", api.AllChirps())
	srvmux.Handle("POST /api/chirps", api.ChirpVal())

	srvmux.Handle("POST /api/polka/webhooks", api.UpgradeUser())

	srvmux.Handle("POST /api/revoke", api.Revoke())
	srvmux.Handle("POST /api/refresh", api.Refreshtoken())

	srvmux.Handle("/app/", http.StripPrefix("/app/", api.MiddlewareHits(http.FileServer(http.Dir(".")))))
	srvmux.Handle("POST /admin/reset", api.ResetHits())
	srvmux.Handle("GET /admin/metrics", api.RetrieveHits())

	srvmux.HandleFunc("GET /api/healthz", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
	})
	Server := &http.Server{Handler: srvmux, Addr: ":8080"}
	Server.ListenAndServe()
}
