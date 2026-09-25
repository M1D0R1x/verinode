package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "healthy",
			Service:   "verinode-api-gateway",
			Version:   "0.1.0",
			Timestamp: time.Now().UTC(),
		})
	})

	mux.HandleFunc("/v1/grades", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		grades := []map[string]interface{}{
			{
				"id":                     "H100-SXM-8XNV",
				"gpu_sku":                "NVIDIA H100 SXM 80GB",
				"min_memory_gb":          640,
				"topology":               "SXM/HGX 8x NVLink 4.0 / NVSwitch",
				"min_healthy_gpu_count":  8,
				"benchmark_floor": map[string]interface{}{
					"nccl_allreduce_gb_per_sec": 400.0,
				},
				"min_cpu_cores": 112,
				"min_ram_gb":    1024,
			},
		}
		_ = json.NewEncoder(w).Encode(grades)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Verinode API Gateway starting on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
