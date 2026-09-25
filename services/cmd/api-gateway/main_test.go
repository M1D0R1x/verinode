package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

func setupTestServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
		})
	})

	mux.HandleFunc("GET /v1/grades", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "H100-SXM-8XNV", "gpu_sku": "NVIDIA H100 SXM 80GB"},
		})
	})

	mux.HandleFunc("POST /v1/contracts/validate-transition", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CurrentState   string                 `json:"current_state"`
			NextState      string                 `json:"next_state"`
			Actor          string                 `json:"actor"`
			Reason         string                 `json:"reason"`
			IdempotencyKey string                 `json:"idempotency_key"`
			AuthDecision   map[string]interface{} `json:"authorization_decision"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
			return
		}
		if req.CurrentState == "settled" && req.NextState == "open" {
			writeProblem(w, http.StatusUnprocessableEntity, "Transition Rejected", "terminal state")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"valid": true})
	})

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
		pubKeyBytes, _ := base64.StdEncoding.DecodeString(req.PublicKey)
		sigBytes, _ := base64.StdEncoding.DecodeString(req.Signature)
		if err := telemetry.VerifyHardwareAttestation(ed25519.PublicKey(pubKeyBytes), req.Report, sigBytes); err != nil {
			writeProblem(w, http.StatusUnauthorized, "Cryptographic Verification Failed", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"verified": true})
	})

	return mux
}

func TestHealthCheckEndpoint(t *testing.T) {
	t.Parallel()

	handler := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", res["status"])
	}
}

func TestGradesEndpoint(t *testing.T) {
	t.Parallel()

	handler := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/grades", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestValidateTransitionEndpoint(t *testing.T) {
	t.Parallel()

	handler := setupTestServer()

	t.Run("Valid transition returns 200", func(t *testing.T) {
		t.Parallel()

		body := map[string]interface{}{
			"current_state":   "open",
			"next_state":      "quoted",
			"actor":           "usr_buyer_1",
			"reason":          "quote received",
			"idempotency_key": "idemp_01",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/contracts/validate-transition", bytes.NewReader(jsonBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("Illegal transition returns 422 Unprocessable Entity", func(t *testing.T) {
		t.Parallel()

		body := map[string]interface{}{
			"current_state":   "settled",
			"next_state":      "open",
			"actor":           "usr_buyer_1",
			"reason":          "reopen",
			"idempotency_key": "idemp_02",
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/contracts/validate-transition", bytes.NewReader(jsonBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
	})
}

func TestTelemetryAttestationVerificationEndpoint(t *testing.T) {
	t.Parallel()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("key generation failed: %v", err)
	}

	report := telemetry.HardwareReport{
		AgentID:   "agent_001",
		BlockID:   "block_001",
		Timestamp: time.Now().UTC(),
		GPUSKU:    "NVIDIA H100 SXM 80GB",
		GPUCount:  8,
		MemoryGB:  640,
		Topology:  "SXM5/HGX NVLink",
	}

	payload, _ := report.CanonicalPayload()
	sig := ed25519.Sign(privKey, payload)

	handler := setupTestServer()

	t.Run("Valid signature returns 200", func(t *testing.T) {
		t.Parallel()

		body := map[string]interface{}{
			"report":     report,
			"signature":  base64.StdEncoding.EncodeToString(sig),
			"public_key": base64.StdEncoding.EncodeToString(pubKey),
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/telemetry/attestations/verify", bytes.NewReader(jsonBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("Tampered report returns 401 Unauthorized", func(t *testing.T) {
		t.Parallel()

		tamperedReport := report
		tamperedReport.GPUCount = 4

		body := map[string]interface{}{
			"report":     tamperedReport,
			"signature":  base64.StdEncoding.EncodeToString(sig),
			"public_key": base64.StdEncoding.EncodeToString(pubKey),
		}
		jsonBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/telemetry/attestations/verify", bytes.NewReader(jsonBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}
