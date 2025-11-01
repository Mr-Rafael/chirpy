package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	port := ":8080"
	mux := http.NewServeMux()

	var config apiConfig
	config.fileserverHits.Store(0)

	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./files")))))
	mux.HandleFunc("/healthz", handlerHealthZ)
	mux.HandleFunc("/metrics", config.handlerMetrics)
	mux.HandleFunc("/reset", config.handlerReset)

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
