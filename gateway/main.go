package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Host struct {
	IP     string
	Status string
}

// Service represents a service with its ID, host, port, and status.
type Service struct {
	ID    string `json:"id"`
	Hosts []Host `json:"hosts"`
	Port  int    `json:"port"`
	// Status string `json:"status"`
	// Add other metadata as needed
}

func discoverServices(discoveredHosts chan<- []*Service) {
	for {
		var serviceResponse []*Service
		call := fmt.Sprintf("http://localhost:8081/discover")

		resp, err := http.Get(call)
		if err != nil {
			log.Println("Failed to discover services (response error)", err)
			time.Sleep(5 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to discover services (read body)", err)
			time.Sleep(5 * time.Second)
			continue
		}
		resp.Body.Close()

		if err := json.Unmarshal(body, &serviceResponse); err != nil {
			log.Println("Failed to discover services (unmarshal body)", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// hosts := serviceResponse[0].Hosts
		// if len(hosts) == 0 {
		// 	log.Println("Failed to discover services (no hosts found)", err)
		// }

		discoveredHosts <- serviceResponse
		time.Sleep(5 * time.Second)
	}

}

func main() {
	discoveredSvc := make(chan []*Service)
	proxy := NewLoadBalancingReverseProxy()

	go func() {
		listenPort := ":8080"
		// use http.Handle instead of http.HandleFunc when your struct implements http.Handler interface
		r := mux.NewRouter()
		r.StrictSlash(true)

		discSvc := <-discoveredSvc
		for _, svc := range discSvc {
			// svcListenString := fmt.Sprintf("/%s/{endpoint}", svc.ID)
			// r.HandleFunc(svcListenString, proxy.ServeHTTP).Name(svc.ID)
			r.PathPrefix("/" + svc.ID + "/{endpoint}").HandlerFunc(proxy.ServeHTTP)
			r.PathPrefix("/" + svc.ID).HandlerFunc(proxy.ServeHTTP)
		}

		log.Printf("Listening on services, port: %s\n", listenPort)

		//spaghetti code
		// func(r *mux.Router, svcs []*Service) []string {
		// 	var routes []string
		// 	for _, svc := range svcs {
		// 		routes = append(routes, r.GetRoute(svc.ID).GetName())
		// 	}
		// 	return routes
		// }(r, discSvc),
		log.Println(http.ListenAndServe(listenPort, r))

	}()

	go discoverServices(discoveredSvc)

	for dscSvc := range discoveredSvc {
		// Update the load balancing reverse proxy with the discovered services
		proxy.UpdateService(dscSvc)
	}

	select {}
}
