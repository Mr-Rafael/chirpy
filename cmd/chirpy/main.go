package main

import (
	"fmt"
	"net/http"
) 

func main() {
	mux := http.NewServeMux()
	port := ":8080"
	
	server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	fmt.Printf("Starting server on %v\n", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server error: %v", err)
	}
}