// Entry point for REST API
package main

import (
	"log"
	"net/http"

	"github.com/checkin-service/internal/api"
)

func main() {
	router := api.NewRouter()

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
