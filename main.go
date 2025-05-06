package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// Create a new storage
	storage := NewStorage()

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a new handler
	handler := NewHandler(storage, []string{"//id", "//@id"}, "X-Resource-ID", logger)

	// Start the server
	http.ListenAndServe(":8080", handler)
}
