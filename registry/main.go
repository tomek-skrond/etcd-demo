package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

// Registry represents a service registry.
type Registry struct {
	services map[string]Service
	mu       sync.RWMutex
}

// NewRegistry creates a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]Service),
	}
}

// Register adds a service to the registry.
func (r *Registry) Register(services ...Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range services {
		r.services[s.ID] = s
		fmt.Printf("Registered service: %s\n", s.ID)
	}
}

// SetStatus updates the status of a service in the registry.
func (r *Registry) SetStatus(serviceID, status, hostIP string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentService, ok := r.services[serviceID]
	if !ok {
		// Service not found
		return
	}

	var updatedHosts []Host
	for _, host := range currentService.Hosts {
		if host.IP == hostIP {
			host.Status = status
		}
		updatedHosts = append(updatedHosts, host)
	}

	// Update the service with the modified hosts slice
	currentService.Hosts = updatedHosts
	r.services[serviceID] = currentService

	log.Printf("Updated service %s host status: %s on ip [%s]:%d\n", serviceID, status, hostIP, currentService.Port)

	// if _, exists := r.services[serviceID]; exists {
	// 	// r.services[serviceID].Status = status
	// 	currentService := r.services[serviceID]
	// 	currentService.Status = status

	// 	r.services[serviceID] = currentService

	// 	log.Printf("Updated status for service %s Status: %s on ip [%s]:%d\n", serviceID, status, currentService.Hosts, currentService.Port)
	// }
}

func containsFieldValue(host Host, value string) (int, bool) {
	fmt.Println(host)
	if host.IP == value {
		return 1, true
	}
	return -1, false
}

// GetServiceByID retrieves a service by its ID.
func (r *Registry) GetServiceByID(id string) ([]Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var services []Service
	serviceFound := false

	for _, s := range r.services {
		if s.ID == id {
			serviceFound = true
			services = append(services, s)
		}
	}
	return services, serviceFound
}

// DiscoverHandler handles requests from clients to discover services.
func DiscoverServiceByID(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		services, ok := registry.GetServiceByID(id)
		if !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(services)
	}
}

func DiscoverAllServices(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var svcs []Service
		for _, value := range registry.services {
			svcs = append(svcs, value)
		}

		json.NewEncoder(w).Encode(svcs)
	}
}

type Config struct {
	serviceConfig ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// func configureService(conf chan<- Config) {
// 	configPath := os.Getenv("CONFIG_PATH")
// 	configData, err := os.ReadFile(configPath)
// 	if err != nil {
// 		log.Fatalln("No config provided")
// 	}
// 	var config []Config
// 	if err := yaml.Unmarshal(configData, &config); err != nil {
// 		log.Fatalln("Bad config")
// 	}

// }
func main() {
	registry := NewRegistry()

	// Example service registration
	service1 := Service{
		ID: "service1",
		Hosts: []Host{
			Host{"192.168.1.88", ""},
			Host{"localhost", ""},
		},
		Port: 7777,
	}

	service2 := Service{
		ID: "service2",
		Hosts: []Host{
			Host{"192.168.1.88", ""},
			Host{"localhost", ""},
		},
		Port: 9999,
	}

	registry.Register(service1, service2)

	// go HealthCheck(registry, 5*time.Second)
	go HealthCheck(registry, 5*time.Second)

	// HTTP endpoints
	http.HandleFunc("/service", DiscoverServiceByID(registry))
	http.HandleFunc("/discover", DiscoverAllServices(registry))

	// Start the HTTP server
	fmt.Println("Service registry running on :8081")
	http.ListenAndServe(":8081", nil)

	select {}
}

func HealthCheck(registry *Registry, interval time.Duration) {
	for {
		registry.mu.RLock()
		for _, service := range registry.services {
			go func(service Service) {
				for _, host := range service.Hosts {
					url := fmt.Sprintf("http://%s:%d/health", host.IP, service.Port)
					resp, err := http.Get(url)
					if err != nil || resp.StatusCode != http.StatusOK {
						registry.SetStatus(service.ID, "inactive", host.IP)
					} else {
						registry.SetStatus(service.ID, "active", host.IP)
					}
				}
			}(service)
		}
		registry.mu.RUnlock()
		time.Sleep(interval)
	}
}
