package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

type LoadBalancingReverseProxy struct {
	proxy    *httputil.ReverseProxy
	services []*Service
}

// NewLoadBalancingReverseProxy creates a new instance of LoadBalancingReverseProxy.
func NewLoadBalancingReverseProxy() *LoadBalancingReverseProxy {
	return &LoadBalancingReverseProxy{
		proxy:    &httputil.ReverseProxy{},
		services: []*Service{}, // Initialize with an empty list of hosts
	}
}

func (lb *LoadBalancingReverseProxy) UpdateService(svc []*Service) {

	lb.services = svc
	for _, s := range lb.services {
		log.Printf("Updated %s hosts: %s\n", s.ID, s.Hosts)
	}

}

// SelectRandomHost selects a random host from the list of available hosts.
func (lb *LoadBalancingReverseProxy) SelectRandomHostOfService(svcId string) *url.URL {
	var activeHosts []Host
	currentServiceIndex := lb.GetCurrentServiceIndex(svcId)
	if currentServiceIndex == -1 {
		log.Println("No such service", svcId)
		return nil
	}
	for _, s := range lb.services {
		if s.ID == svcId {
			for _, host := range s.Hosts {
				if host.Status == "active" {
					activeHosts = append(activeHosts, host)
				}
			}
		}
	}
	// fmt.Printf("active hosts of service %s: %s\n", svcId, activeHosts)
	if len(activeHosts) > 0 {
		selectedHost := activeHosts[rand.Intn(len(activeHosts))]
		hostAndPort := fmt.Sprintf("%s:%d", selectedHost.IP, lb.services[currentServiceIndex].Port)
		return &url.URL{
			Scheme: "http",
			Host:   hostAndPort,
		}
	}
	return nil
}

func (lb *LoadBalancingReverseProxy) GetCurrentServiceIndex(svcId string) int {
	// fmt.Println(svcId)
	// fmt.Println(lb.services)
	for i, s := range lb.services {
		// fmt.Println("service id:", s.ID)
		if svcId == s.ID {
			return i
		}
	}
	return -1
}

// ServeHTTP forwards the HTTP request to one of the available hosts using load balancing.
func (lb *LoadBalancingReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srcHost := r.Host
	srcScheme := "http://"
	srcPath := r.URL.Path

	// whole path example: http://localhost:8080/service1/health
	// should get rewritten into http://<service_host>:<service_port>/health
	// targetServiceRoute is a sub path that is equal to destination url (for ex. /health or / or /service)
	var targetServiceRoute string
	endpoint := mux.Vars(r)["endpoint"]

	wholePath := r.URL.Path
	currentServiceId := getCurrentServiceFromURLPath(wholePath)

	if endpoint == "" || endpoint == "/" {
		targetServiceRoute = "/"
	} else {
		targetServiceRoute = "/" + endpoint
	}
	// fmt.Println(targetServiceRoute, srcHost)

	currentServiceIndex := lb.GetCurrentServiceIndex(currentServiceId)
	if currentServiceIndex == -1 {
		http.Error(w, "No such service available", http.StatusServiceUnavailable)
		return
	}
	targetURL := lb.SelectRandomHostOfService(currentServiceId)
	// fmt.Printf("SERVICE %s URL %s\n", currentServiceId, targetURL)
	if targetURL == nil {
		msg := fmt.Sprintf("No active hosts in service %s available", currentServiceId)
		http.Error(w, msg, http.StatusServiceUnavailable)
		return
	}

	lb.proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = targetServiceRoute
	}

	if !isPublic(targetServiceRoute) {
		authHeader := r.Header.Get("Authorization")

		if authenticate(authHeader) {
			log.Println(r.Method, srcScheme+srcHost+srcPath, "->", targetURL.String()+targetServiceRoute)
			// w.Header().Set("X-Forwarded-For", targetURL.Host)
			lb.proxy.ServeHTTP(w, r)
		} else {
			log.Println(r.Method, srcScheme+srcHost+srcPath, "->", targetURL.String()+targetServiceRoute)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
		}
	} else {
		log.Println(r.Method, srcScheme+srcHost+srcPath, "->", targetURL.String()+targetServiceRoute)
		// w.Header().Set("X-Forwarded-For", targetURL.Host)
		lb.proxy.ServeHTTP(w, r)
	}
}

func getCurrentServiceFromURLPath(wholePath string) string {
	var urlparts []string
	for _, part := range strings.Split(wholePath, "/") {
		if len(part) != 0 {
			urlparts = append(urlparts, part)
		}
	}
	currentServie := urlparts[0]
	return currentServie
}

func isPublic(path string) bool {
	if path == "/" {
		return true
	}
	if path == "/service" {
		return false
	}
	return false
}

func authenticate(authHeader string) bool {
	if authHeader == "" {
		return false
	}
	if authHeader == "validToken" {
		return true
	}
	return false
}
