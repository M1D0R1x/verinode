package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/auth"
	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

func setupTestServer() http.Handler {
	return NewServer(nil, nil, nil, nil, nil, nil, nil, nil)
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

func TestUnconfiguredDatabaseGracefulServiceUnavailable(t *testing.T) {
	t.Parallel()

	// Server with nil repos returns 503 Service Unavailable with RFC 7807 problem details
	handler := setupTestServer()

	endpoints := []struct {
		method  string
		path    string
		body    string
		needsAuth bool
	}{
		{"POST", "/v1/participants", `{"legal_name":"Test LLC"}`, false},
		{"GET", "/v1/participants", "", false},
		{"POST", "/v1/inventory/blocks", `{"grade_id":"H100-SXM-8XNV"}`, true},
		{"GET", "/v1/inventory/blocks", "", false},
		{"POST", "/v1/rfqs", `{"grade_id":"H100-SXM-8XNV"}`, true},
		{"GET", "/v1/rfqs/some-id", "", false},
		{"POST", "/v1/rfqs/some-id/quotes", `{"price_cents":100}`, true},
		{"GET", "/v1/rfqs/some-id/quotes", "", false},
		{"POST", "/v1/rfqs/some-id/quotes/q-id/accept", `{"buyer_id":"b-id"}`, true},
		{"GET", "/v1/contracts/c-id", "", false},
		{"POST", "/v1/contracts/c-id/advance", `{"next_state":"live"}`, false},
	}

	// A valid trader token so trading-write endpoints pass auth and reach the DB check.
	traderToken, _ := auth.NewTokenService(os.Getenv("AUTH_JWT_SECRET"), time.Hour).
		Mint("usr_test", "t@co.example", auth.RoleTrader, "co_test")

	for _, ep := range endpoints {
		ep := ep
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			t.Parallel()
			var req *http.Request
			if ep.body != "" {
				req = httptest.NewRequest(ep.method, ep.path, bytes.NewBufferString(ep.body))
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}
			if ep.needsAuth {
				req.Header.Set("Authorization", "Bearer "+traderToken)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503 Service Unavailable, got %d for %s %s", rec.Code, ep.method, ep.path)
			}
		})
	}
}

