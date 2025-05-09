// This is only meant for local development
package main

import (
	"fmt"
	"net/http"

	handler "github.com/Ankitz007/leetlink/api"
)

func main() {
	// Register the Handler function to the default router
	http.HandleFunc("/api/leetlink/", handler.Handler)
	http.HandleFunc("/api/cron/", handler.Cron)

	// Start the HTTP server
	// Note: ":8080" is the port number; you can choose a different one if needed.
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
