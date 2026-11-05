package main

import (
	"log"
	"net/http"
)

const (
	port = "8080"
)

func main() {
	handler := http.NewServeMux()

	log.Printf("Server is running on port=%v", port)
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatal("failed to start server")
	}
}
