package arbitrum

import (
	"context"
	"log/slog"
	"math/big"
	"testing"
)

func TestArbitrumAdapter_Invariants(t *testing.T) {
	adapter := NewArbitrumAdapter("", "", slog.Default())
	ctx := context.Background()

	tradeUUID := "c37610f3-4aea-414f-868c-5937e97e822d"
	tradeBytes, err := adapter.ComputeTradeBytes32(tradeUUID)
	if err != nil {
		t.Fatalf("ComputeTradeBytes32 failed: %v", err)
	}

	t.Run("MirrorTradeToArbitrum produces valid txHash", func(t *testing.T) {
		env := TradeEnvelopeEVM{
			TradeID:         tradeBytes,
			BuyerAddress:    "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			SellerAddress:   "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
			EscrowAmountWei: big.NewInt(1000000000000000000), // 1 ETH
			WindowStart:     1727280000,
			WindowEnd:       1727884800,
			State:           1, // FundedSecured
		}

		txHash, err := adapter.MirrorTradeToArbitrum(ctx, env)
		if err != nil {
			t.Fatalf("MirrorTradeToArbitrum failed: %v", err)
		}
		if len(txHash) == 0 {
			t.Errorf("Expected non-empty txHash")
		}
	})

	t.Run("Attestation below benchmark floor (<400 GB/s) is rejected", func(t *testing.T) {
		var digest [32]byte
		_, err := adapter.MirrorAttestationToArbitrum(ctx, tradeBytes, digest, 340, true)
		if err == nil {
			t.Errorf("Expected error when NCCL bandwidth is below 400 GB/s, got nil")
		}
	})

	t.Run("Compliant attestation (>=400 GB/s) succeeds", func(t *testing.T) {
		var digest [32]byte
		txHash, err := adapter.MirrorAttestationToArbitrum(ctx, tradeBytes, digest, 412, true)
		if err != nil {
			t.Fatalf("Expected success for compliant attestation: %v", err)
		}
		if len(txHash) == 0 {
			t.Errorf("Expected non-empty txHash")
		}
	})
}
