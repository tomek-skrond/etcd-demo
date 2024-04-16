package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	listenPort := ":9999"

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/",
			"instance":         "A",
			"port":             listenPort,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())
	})
	r.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/service",
			"instance":         "A",
			"port":             listenPort,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())
	})
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/health",
			"instance":         "A",
			"port":             listenPort,
			"status":           "healthy",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())

	})
	log.Fatalln(http.ListenAndServe(listenPort, r))
}
