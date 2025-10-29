package main

import (
	"log"
	"fmt"
	"net/http"
) 

func main() {
	port := ":8080"
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./files"))))
	mux.HandleFunc("/healthz", handlerHealthZ)
	
	server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	fmt.Printf("Starting server on %v\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    	log.Fatalf("server error: %v", err)
	}
}