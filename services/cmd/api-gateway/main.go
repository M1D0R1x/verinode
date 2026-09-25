package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	TraceID  string `json:"trace_id,omitempty"`
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ProblemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// 1. Health Probe
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "verinode-api-gateway",
			"version":   "0.1.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// 2. Delivery Grades
	mux.HandleFunc("GET /v1/grades", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		grades := []map[string]interface{}{
			{
				"id":                    "H100-SXM-8XNV",
				"gpu_sku":               "NVIDIA H100 SXM 80GB",
				"min_memory_gb":         640,
				"topology":              "SXM5/HGX 8x NVLink 4.0 / NVSwitch (900 GB/s bidirectional)",
				"min_healthy_gpu_count": 8,
				"benchmark_floor": map[string]interface{}{
					"nccl_allreduce_gb_per_sec":   400.0,
					"min_cuda_driver":             "535.129.03",
					"max_ecc_unrecovered_errors": 0,
				},
				"min_cpu_cores": 112,
				"min_ram_gb":    1024,
				"min_nvme_perf": 100000,
			},
		}
		_ = json.NewEncoder(w).Encode(grades)
	})

	// 3. Contract State Transition Validator (Phase 1 Engine Endpoint)
	mux.HandleFunc("POST /v1/contracts/validate-transition", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CurrentState   string                   `json:"current_state"`
			NextState      string                   `json:"next_state"`
			Actor          string                   `json:"actor"`
			Reason         string                   `json:"reason"`
			IdempotencyKey string                   `json:"idempotency_key"`
			AuthDecision   map[string]interface{}   `json:"authorization_decision"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
			return
		}

		current, err := contract.ParseState(req.CurrentState)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Invalid Current State", err.Error())
			return
		}

		next, err := contract.ParseState(req.NextState)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Invalid Target State", err.Error())
			return
		}

		evt := contract.TransitionEvent{
			Actor:                 req.Actor,
			Reason:                req.Reason,
			IdempotencyKey:        req.IdempotencyKey,
			AuthorizationDecision: req.AuthDecision,
		}

		if err := contract.ValidateTransition(current, next, evt); err != nil {
			status := http.StatusUnprocessableEntity
			if errors.Is(err, contract.ErrMissingActor) ||
				errors.Is(err, contract.ErrMissingReason) ||
				errors.Is(err, contract.ErrMissingIdempotency) {
				status = http.StatusBadRequest
			}
			writeProblem(w, status, "Transition Rejected", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":         true,
			"current_state": current,
			"next_state":    next,
			"timestamp":     time.Now().UTC(),
		})
	})

	// 4. Telemetry Attestation Ingestion Verification
	mux.HandleFunc("POST /v1/telemetry/attestations/verify", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Report    telemetry.HardwareReport `json:"report"`
			Signature string                   `json:"signature"`
			PublicKey string                   `json:"public_key"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
			return
		}

		pubKeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
		if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
			writeProblem(w, http.StatusBadRequest, "Invalid Public Key", "Public key must be base64-encoded ed25519 key")
			return
		}

		sigBytes, err := base64.StdEncoding.DecodeString(req.Signature)
		if err != nil || len(sigBytes) != ed25519.SignatureSize {
			writeProblem(w, http.StatusBadRequest, "Invalid Signature", "Signature must be base64-encoded ed25519 signature")
			return
		}

		if err := telemetry.VerifyHardwareAttestation(ed25519.PublicKey(pubKeyBytes), req.Report, sigBytes); err != nil {
			writeProblem(w, http.StatusUnauthorized, "Cryptographic Verification Failed", err.Error())
			return
		}

		if err := telemetry.ValidateAgainstGrade(req.Report, 8, 640, "NVLink"); err != nil {
			writeProblem(w, http.StatusUnprocessableEntity, "Hardware Grade Compliance Failed", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"verified":        true,
			"agent_id":        req.Report.AgentID,
			"block_id":        req.Report.BlockID,
			"grade_compliant": true,
			"timestamp":       time.Now().UTC(),
		})
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Verinode API Gateway initialized", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
