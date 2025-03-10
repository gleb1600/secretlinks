package main

import (
	"log"
	"net/http"
	"secretlinks/handlers"
	"secretlinks/middleware"
	"secretlinks/storage"
)

func main() {
	storage := storage.NewMemoryStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/create", handlers.CreateHandler(storage))
	mux.HandleFunc("/", handlers.RedirectHandler(storage))

	newMux := middleware.LoggingMiddleware(mux)

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", newMux)
}

// Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/" -Body @{secret="i love nika"}
// curl http://127.0.0.1:8080/qELIuIRM
