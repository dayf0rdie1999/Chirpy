package main

import (
	"database/sql"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/dayf0rdie1999/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbUrl)

	mux := http.NewServeMux()
	config := apiConfig{
		fileserverHits: atomic.Int32{},
		queries:        database.New(db),
		authSecret:     os.Getenv("AUTH_SECRET"),
		polkaApiKey:    os.Getenv("POLKA_API_KEY"),
	}
	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(200)
		res.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /admin/reset", config.resetMetric)
	mux.HandleFunc("GET /admin/metrics", config.renderAdminMetrics)
	mux.HandleFunc("POST /api/users", config.handleCreateUser)
	mux.HandleFunc("POST /api/chirps", config.handleInsertChirps)
	mux.HandleFunc("GET /api/chirps", config.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.getChirpById)
	mux.HandleFunc("POST /api/login", config.handleUserLogin)
	mux.HandleFunc("POST /api/refresh", config.refreshAccessToken)
	mux.HandleFunc("POST /api/revoke", config.RevokeRefreshToken)
	mux.HandleFunc("PUT /api/users", config.UpdateUsers)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", config.DeleteChrips)
	mux.HandleFunc("POST /api/polka/webhooks", config.UpgradeMembership)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
