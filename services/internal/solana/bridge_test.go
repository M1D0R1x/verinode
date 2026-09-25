package solana_test

import (
	"context"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/solana"
)

func TestSolanaMirrorAdapter_Invariants(t *testing.T) {
	t.Parallel()

	adapter := solana.NewMirrorAdapter("", "", nil)

	env := solana.TradeEnvelopeAccount{
		TradeID:     "c8f921a4-9b2e-4b13-91dc-837264819011",
		BuyerPubkey: "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
		GradeID:     "H100-SXM-8XNV",
		State:       "funded_secured",
		PriceCents:  3528000,
		WindowStart: time.Now().UTC(),
		WindowEnd:   time.Now().UTC().Add(168 * time.Hour),
	}

	ctx := context.Background()

	t.Run("MirrorTradeEnvelope produces devnet transaction", func(t *testing.T) {
		t.Parallel()
		sig, err := adapter.MirrorTradeEnvelope(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error mirroring trade envelope: %v", err)
		}
		if len(sig) == 0 {
			t.Errorf("expected non-empty transaction signature")
		}
	})

	t.Run("Attestation below benchmark floor is rejected", func(t *testing.T) {
		t.Parallel()
		var digest [32]byte
		// 350 GB/s is below the 400 GB/s floor for benchmark grade H100-SXM-8XNV
		_, err := adapter.MirrorAttestationProof(ctx, env.TradeID, digest, 350, true)
		if err == nil {
			t.Fatalf("expected attestation below 400 GB/s to be rejected by floor check, but passed!")
		}
	})

	t.Run("Compliant attestation succeeds", func(t *testing.T) {
		t.Parallel()
		var digest [32]byte
		sig, err := adapter.MirrorAttestationProof(ctx, env.TradeID, digest, 428, true)
		if err != nil {
			t.Fatalf("unexpected error on compliant attestation: %v", err)
		}
		if len(sig) == 0 {
			t.Errorf("expected non-empty transaction signature")
		}
	})
}
