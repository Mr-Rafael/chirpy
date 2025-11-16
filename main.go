package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Mr-Rafael/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	port := ":8080"
	mux := http.NewServeMux()
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening connection to the database: %v", err)
	}

	var config apiConfig
	config.fileserverHits.Store(0)
	config.db = database.New(db)
	config.platform = os.Getenv("PLATFORM")
	config.secret = os.Getenv("SECRET")

	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./files")))))
	mux.HandleFunc("GET /api/healthz", handlerHealthZ)
	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerReset)
	mux.HandleFunc("POST /api/users", config.handlerUsers)
	mux.HandleFunc("POST /api/chirps", config.handlerChirpsPOST)
	mux.HandleFunc("GET /api/chirps", config.handlerChirpsGET)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", config.handlerChirpsGETID)
	mux.HandleFunc("POST /api/login", config.handlerLogin)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	fmt.Printf("Starting server on %v\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
