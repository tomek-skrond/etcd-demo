package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	listenPort := ":7777"

	instance, err := os.Hostname()
	if err != nil {
		instance = fmt.Sprintf("%d", rand.Intn(1000))
	}

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/",
			"instance":         instance,
			"port":             listenPort,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())
	})
	r.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/service",
			"instance":         instance,
			"port":             listenPort,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())
	})
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service_endpoint": "/health",
			"instance":         instance,
			"port":             listenPort,
			"status":           "healthy",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("%s %s %s", r.URL, r.Header["X-Forwarded-For"], r.UserAgent())

	})
	log.Fatalln(http.ListenAndServe(listenPort, r))
}
