package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalancingReverseProxy struct {
	proxy   *httputil.ReverseProxy
	service Service
}

// NewLoadBalancingReverseProxy creates a new instance of LoadBalancingReverseProxy.
func NewLoadBalancingReverseProxy() *LoadBalancingReverseProxy {
	return &LoadBalancingReverseProxy{
		proxy:   &httputil.ReverseProxy{},
		service: Service{}, // Initialize with an empty list of hosts
	}
}

func (lb *LoadBalancingReverseProxy) UpdateService(svc Service) {
	lb.service = svc
}

// SelectRandomHost selects a random host from the list of available hosts.
func (lb *LoadBalancingReverseProxy) SelectRandomHost() *url.URL {
	var activeHosts []Host
	for _, host := range lb.service.Hosts {
		if host.Status == "active" {
			activeHosts = append(activeHosts, host)
		}
	}

	if len(activeHosts) > 0 {
		selectedHost := activeHosts[rand.Intn(len(activeHosts))]
		hostAndPort := fmt.Sprintf("%s:%d", selectedHost.IP, lb.service.Port)
		return &url.URL{
			Scheme: "http",
			Host:   hostAndPort,
		}
	}
	return nil
}

// ServeHTTP forwards the HTTP request to one of the available hosts using load balancing.
func (lb *LoadBalancingReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srcHost := r.Host
	// fmt.Println(srcHost)

	targetURL := lb.SelectRandomHost()
	if targetURL == nil {
		http.Error(w, "No active hosts available", http.StatusServiceUnavailable)
		return
	}
	lb.proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = r.URL.Path
	}

	targetHost := targetURL.Host
	path := r.URL.Path

	// fmt.Println(targetHost)
	// fmt.Println(path)

	if !isPublic(path) {
		authHeader := r.Header.Get("Authorization")

		if authenticate(authHeader) {
			log.Println(r.Method, srcHost, "->", targetURL)
			w.Header().Set("X-Forwarded-For", targetHost)
			lb.proxy.ServeHTTP(w, r)
		} else {
			log.Println(r.Method, srcHost, "->", targetURL)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
		}
	} else {
		log.Println(r.Method, srcHost, "->", targetURL)
		w.Header().Set("X-Forwarded-For", targetHost)
		lb.proxy.ServeHTTP(w, r)
	}
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

// func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

// 	host := r.Host
// 	path := r.URL.Path
// 	// qs := r.URL.Query()

// 	if !isPublic(path) {
// 		authHeader := r.Header.Get("Authorization")
// 		if authenticate(authHeader) {
// 			log.Println(r.URL, host, "->", ph.URL.Host)
// 			r.Host = ph.URL.Host
// 			w.Header().Set("X-Forwarded-For", host)
// 			ph.p.ServeHTTP(w, r)
// 		} else {
// 			http.Error(w, "Not authorized", http.StatusUnauthorized)
// 		}
// 	} else {
// 		log.Println(r.URL, host, "->", ph.URL.Host)
// 		r.Host = ph.URL.Host
// 		w.Header().Set("X-Forwarded-For", host)
// 		ph.p.ServeHTTP(w, r)
// 	}
// }
