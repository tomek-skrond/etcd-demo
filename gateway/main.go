package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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

func discoverServices(svc string, discoveredHosts chan<- Service) {
	for {
		var serviceResponse []Service
		call := fmt.Sprintf("http://localhost:8081/discover?id=%s", svc)

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

		discoveredHosts <- serviceResponse[0]
		time.Sleep(5 * time.Second)
	}

}

func main() {
	discoveredSvc := make(chan Service)
	proxy := NewLoadBalancingReverseProxy()

	go func() {
		// use http.Handle instead of http.HandleFunc when your struct implements http.Handler interface
		// r := mux.NewRouter()

		http.HandleFunc("/", proxy.ServeHTTP)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	go discoverServices("service1", discoveredSvc)

	for dscSvc := range discoveredSvc {
		// dscSvc := <-discoveredSvc
		// Update the load balancing reverse proxy with the discovered services
		proxy.UpdateService(dscSvc)
		log.Println("Updated hosts:", dscSvc)
		// time.Sleep(2 * time.Second)

	}

	select {}
}
