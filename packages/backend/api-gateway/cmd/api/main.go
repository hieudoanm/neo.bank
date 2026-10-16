package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router := api.NewRouter()
	log.Printf("api-gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
