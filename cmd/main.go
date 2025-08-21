package main

import (
	"fmt"
	"hot-coffee1/internal/config"
	"hot-coffee1/internal/handler"
	"log"
	"net/http"
)

func main() {
	if err := config.ConfigLoad(); err != nil {
		log.Fatal(err)
	}

	port := config.GetConfigPort()

	mux := http.NewServeMux()

	handler.InventoryEndpoints(mux)
	handler.MenuEndpoints(mux)
	handler.OrderEndpoints(mux)
	handler.AggregationEndpoints(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.ErrorResponse(w, "405 - No such method", http.StatusMethodNotAllowed)
	})

	fmt.Println("Server started listening on port -", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
