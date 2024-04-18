package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func discoverServices(discoveredHosts chan<- []*Service, interval time.Duration) {
	for {
		registryHostname := os.Getenv("REGISTRY_HOSTNAME")
		if registryHostname == "" {
			registryHostname = "localhost"
		}
		var serviceResponse []*Service
		call := fmt.Sprintf("http://%s:8081/discover", registryHostname)

		resp, err := http.Get(call)
		if err != nil {
			log.Println("Failed to discover services (response error)", err)
			time.Sleep(interval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to discover services (read body)", err)
			time.Sleep(interval)
			continue
		}
		resp.Body.Close()

		if err := json.Unmarshal(body, &serviceResponse); err != nil {
			log.Println("Failed to discover services (unmarshal body)", err)
			time.Sleep(interval)
			continue
		}

		discoveredHosts <- serviceResponse
		time.Sleep(interval)
	}

}

func main() {

	polling_rate := flag.Duration("polling-rate", 1000*time.Millisecond, "Interval for polling service discovery [ms]")
	flag.Parse()

	discoveredSvc := make(chan []*Service)
	proxy := NewLoadBalancingReverseProxy()

	go func() {
		listenPort := ":8080"
		r := http.NewServeMux()
		// r.StrictSlash(true)

		discSvc := <-discoveredSvc
		for _, svc := range discSvc {
			// svcListenString := fmt.Sprintf("/%s/{endpoint}", svc.ID)
			// r.HandleFunc(svcListenString, proxy.ServeHTTP).Name(svc.ID)
			r.HandleFunc("/"+svc.ID+"/{endpoint...}", proxy.ServeHTTP)
			r.HandleFunc("/"+svc.ID, proxy.ServeHTTP)
		}

		log.Printf("Listening on services, port: %s\n", listenPort)
		log.Println(http.ListenAndServe(listenPort, r))

	}()

	go discoverServices(discoveredSvc, *polling_rate)

	for dscSvc := range discoveredSvc {
		// Update the load balancing reverse proxy with the discovered services
		proxy.UpdateService(dscSvc)
	}

	select {}
}
