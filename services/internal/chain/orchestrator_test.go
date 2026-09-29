package chain

import (
	"context"
	"strings"
	"testing"
	"time"
)

func sampleEnvelope() TradeEnvelope {
	return TradeEnvelope{
		TradeID:      "c37610f3-4aea-414f-868c-5937e97e822d",
		BuyerID:      "buyer-anthropic-spv",
		SellerID:     "seller-lambda-labs",
		GradeID:      "H100-SXM-8XNV",
		State:        "scheduled",
		PriceCents:   2956800,
		WindowStart:  time.Now().UTC(),
		WindowEnd:    time.Now().UTC().Add(168 * time.Hour),
		NCCLGbps:     405,
		CanaryPassed: true,
	}
}

func TestOrchestrator_MirrorTrade_SimulatedProofs(t *testing.T) {
	o := NewOrchestrator(Config{}, nil)
	bundle := o.MirrorTrade(context.Background(), sampleEnvelope())

	if len(bundle.Rails) != 2 {
		t.Fatalf("expected 2 rails (solana, arbitrum), got %d", len(bundle.Rails))
	}
	for _, r := range bundle.Rails {
		if r.Error != "" {
			t.Errorf("rail %s reported error: %s", r.Chain, r.Error)
		}
		if r.StateTx == "" {
			t.Errorf("rail %s produced no state tx", r.Chain)
		}
		if r.AttestTx == "" {
			t.Errorf("rail %s produced no attestation tx for a passing canary", r.Chain)
		}
		if r.Mode != "simulated" {
			t.Errorf("expected simulated mode without keys, rail %s got %q", r.Chain, r.Mode)
		}
	}
}

func TestOrchestrator_Modes(t *testing.T) {
	o := NewOrchestrator(Config{}, nil)
	modes := o.Modes()
	for _, chain := range []string{"solana", "arbitrum", "hyperliquid"} {
		if modes[chain] == "" {
			t.Errorf("missing mode for %s", chain)
		}
	}
}

func TestOrchestrator_Hedge_ShortWhenFixedAboveIndex(t *testing.T) {
	o := NewOrchestrator(Config{}, nil)
	// 168h, 8 GPUs, fixed 2.20 vs index 2.10 -> fixed above index -> SHORT_PERP.
	q := o.Hedge(168, 8, 2.20, 2.10)
	if q.TotalGPUHours != 1344 {
		t.Errorf("expected 1344 GPU-hours, got %d", q.TotalGPUHours)
	}
	if q.RecommendedAction != "SHORT_PERP" {
		t.Errorf("expected SHORT_PERP, got %s", q.RecommendedAction)
	}
	if q.BasisSpreadPct <= 0 {
		t.Errorf("expected positive basis pct, got %.2f", q.BasisSpreadPct)
	}
	if !strings.Contains(strings.ToLower(q.Rationale), "short") {
		t.Errorf("rationale should explain the short: %q", q.Rationale)
	}
}

func TestOrchestrator_Hedge_NoHedgeWhenBasisTight(t *testing.T) {
	o := NewOrchestrator(Config{}, nil)
	// fixed 2.201 vs index 2.20 -> basis < 0.5% -> NO_HEDGE.
	q := o.Hedge(168, 8, 2.201, 2.20)
	if q.RecommendedAction != "NO_HEDGE" {
		t.Errorf("expected NO_HEDGE for tight basis, got %s (%.4f%%)", q.RecommendedAction, q.BasisSpreadPct)
	}
}
