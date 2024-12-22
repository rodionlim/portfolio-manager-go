package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Define the directory to serve files from
	uiDir := "./ui/build"

	// Check if the directory exists
	if _, err := os.Stat(uiDir); os.IsNotExist(err) {
		log.Fatalf("UI directory %s does not exist", uiDir)
	}

	// Serve static files
	http.Handle("/", http.FileServer(http.Dir(uiDir)))

	// Start the server
	port := ":8080"
	log.Printf("Serving UI on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
