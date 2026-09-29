package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/claims"
	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/db"
	"github.com/M1D0R1x/verinode/services/internal/inventory"
	"github.com/M1D0R1x/verinode/services/internal/ledger"
	"github.com/M1D0R1x/verinode/services/internal/marketdata"
	"github.com/M1D0R1x/verinode/services/internal/memrepo"
	"github.com/M1D0R1x/verinode/services/internal/participant"
	"github.com/M1D0R1x/verinode/services/internal/rfq"
)

type ProblemDetails struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Detail  string `json:"detail"`
	TraceID string `json:"trace_id,omitempty"`
}

// loadDotEnv loads KEY=VALUE lines from candidate env files WITHOUT overriding
// variables already present in the process environment. Dependency-free. This is why
// running the gateway from anywhere picks up services/.env (e.g. the Neon DATABASE_URL)
// with no shell export step.
func loadDotEnv() {
	for _, path := range []string{".env", "services/.env", "../.env", "../../.env"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			eq := strings.IndexByte(line, '=')
			if eq <= 0 {
				continue
			}
			key := strings.TrimSpace(line[:eq])
			val := strings.TrimSpace(line[eq+1:])
			if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
				val = val[1 : len(val)-1]
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
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

// allowedOrigins is the CORS allowlist. Credentialed requests (cookies) forbid the
// wildcard origin, so we reflect a specific allowlisted origin instead. Override via
// CORS_ALLOWED_ORIGINS (comma-separated) in other environments.
func allowedOrigins() map[string]bool {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:3000,http://127.0.0.1:3000"
	}
	m := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			m[o] = true
		}
	}
	return m
}

var corsAllowlist = allowedOrigins()

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Reflect an allowlisted origin (required: cannot use "*" with credentials).
		if origin != "" && corsAllowlist[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		} else if origin == "" {
			// Non-browser / same-origin callers (curl, server-to-server): permissive,
			// but WITHOUT credentials so no cookie is ever exposed cross-origin.
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Idempotency-Key")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	loadDotEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	var (
		pRepo      ParticipantStore
		iRepo      InventoryStore
		rRepo      RFQStore
		cRepo      ContractStore
		lRepo      *ledger.Repository
		claimRepo  ClaimsStore
		marketRepo MarketStore
	)

	dbCfg := db.DefaultConfig()
	connected := false
	if dbCfg.ConnString != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err := db.Connect(ctx, dbCfg, logger)
		cancel()

		if err != nil {
			logger.Warn("Failed to connect to PostgreSQL; falling back to in-memory store", "error", err)
		} else {
			defer pool.Close()
			pRepo = participant.NewRepository(pool.Pool)
			iRepo = inventory.NewRepository(pool.Pool)
			rRepo = rfq.NewRepository(pool.Pool)
			cRepo = contract.NewRepository(pool.Pool)
			lRepo = ledger.NewRepository(pool.Pool)
			claimRepo = claims.NewRepository(pool.Pool)
			marketRepo = marketdata.NewRepository(pool.Pool)
			connected = true
			logger.Info("All database domain repositories connected (PostgreSQL)")
		}
	}

	if !connected {
		// No database configured or reachable — run fully on seeded in-memory
		// repositories so every screen and API works with zero infrastructure.
		pRepo = memrepo.NewParticipantRepo()
		iRepo = memrepo.NewInventoryRepo()
		rRepo = memrepo.NewRFQRepo()
		cRepo = memrepo.NewContractRepo()
		claimRepo = memrepo.NewClaimsRepo()
		marketRepo = memrepo.NewMarketRepo()
		logger.Warn("Running with the IN-MEMORY store (no DATABASE_URL). Data is seeded and non-persistent.")
	}

	srv := NewServer(pRepo, iRepo, rRepo, cRepo, lRepo, claimRepo, marketRepo, logger)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Info("Verinode API Gateway initialized", "port", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
