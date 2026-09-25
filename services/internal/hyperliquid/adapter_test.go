package hyperliquid

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestHyperliquidAdapter_Invariants(t *testing.T) {
	adapter := NewAdapter("", "", slog.Default())
	ctx := context.Background()

	t.Run("Compliant index fix publishes successfully to Hyperliquid HIP-3 oracle", func(t *testing.T) {
		update := HIP3OracleUpdate{
			OracleID:           "VERINODE-H100",
			Market:             "H100-168H-PERP",
			MarkPriceUSD:       2.25,
			ConfidenceLowUSD:   2.15,
			ConfidenceHighUSD:  2.35,
			ContributorCount:   4,
			TotalNotionalUSD:   250000.0,
			InsufficientData:   false,
			MethodologyVersion: "v1.0.0-vw-median",
			PublishedAt:        time.Now().UTC(),
		}

		txSig, err := adapter.PublishHIP3OracleFeed(ctx, update)
		if err != nil {
			t.Fatalf("Expected publication to succeed: %v", err)
		}
		if len(txSig) == 0 {
			t.Errorf("Expected non-empty transaction signature")
		}
	})

	t.Run("Invariant 5: insufficient_data == true must be rejected from Hyperliquid", func(t *testing.T) {
		update := HIP3OracleUpdate{
			OracleID:         "VERINODE-H100",
			Market:           "H100-168H-PERP",
			MarkPriceUSD:     0,
			ContributorCount: 1,
			InsufficientData: true,
		}

		_, err := adapter.PublishHIP3OracleFeed(ctx, update)
		if err == nil {
			t.Fatalf("Expected error rejecting insufficient data, got nil")
		}
	})

	t.Run("Minimum contributor threshold (<3) must be rejected", func(t *testing.T) {
		update := HIP3OracleUpdate{
			OracleID:         "VERINODE-H100",
			Market:           "H100-168H-PERP",
			MarkPriceUSD:     2.20,
			ContributorCount: 2, // Less than 3
			InsufficientData: false,
		}

		_, err := adapter.PublishHIP3OracleFeed(ctx, update)
		if err == nil {
			t.Fatalf("Expected error rejecting < 3 contributors, got nil")
		}
	})

	t.Run("CalculateHedgeQuote math determinism", func(t *testing.T) {
		// 168 hours, 8 GPUs -> 1344 GPU-hours
		// Fixed rate $2.20 vs mark $2.10
		quote := adapter.CalculateHedgeQuote(168, 8, 2.20, 2.10)
		if quote.TotalGPUHours != 1344 {
			t.Errorf("Expected 1344 total GPU hours, got %d", quote.TotalGPUHours)
		}
		if quote.RecommendedAction != "SHORT_PERP" {
			t.Errorf("Expected SHORT_PERP, got %s", quote.RecommendedAction)
		}
		if quote.BasisSpreadUSD <= 0 {
			t.Errorf("Expected positive basis spread, got %.2f", quote.BasisSpreadUSD)
		}
	})
}
