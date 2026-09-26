package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/claims"
	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/db"
	"github.com/M1D0R1x/verinode/services/internal/inventory"
	"github.com/M1D0R1x/verinode/services/internal/ledger"
	"github.com/M1D0R1x/verinode/services/internal/marketdata"
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Idempotency-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	var (
		pRepo      *participant.Repository
		iRepo      *inventory.Repository
		rRepo      *rfq.Repository
		cRepo      *contract.Repository
		lRepo      *ledger.Repository
		claimRepo  *claims.Repository
		marketRepo *marketdata.Repository
	)

	dbCfg := db.DefaultConfig()
	if dbCfg.ConnString != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err := db.Connect(ctx, dbCfg, logger)
		cancel()

		if err != nil {
			logger.Warn("Failed to connect to PostgreSQL database; running with limited endpoints", "error", err)
		} else {
			defer pool.Close()
			pRepo = participant.NewRepository(pool.Pool)
			iRepo = inventory.NewRepository(pool.Pool)
			rRepo = rfq.NewRepository(pool.Pool)
			cRepo = contract.NewRepository(pool.Pool)
			lRepo = ledger.NewRepository(pool.Pool)
			claimRepo = claims.NewRepository(pool.Pool)
			marketRepo = marketdata.NewRepository(pool.Pool)
			logger.Info("All database domain repositories connected successfully")
		}
	} else {
		logger.Warn("DATABASE_URL not configured; running in standalone mode")
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
