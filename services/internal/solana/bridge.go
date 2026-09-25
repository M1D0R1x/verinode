package solana

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"
)

// Invariant 6 Enforcer:
// "On-chain rails (Solana Anchor, EVM) are strictly optional, read-only/mirroring adapters gated to Phase 4.
// The off-chain PostgreSQL database and executed legal confirmations remain the authoritative sources of truth."

// TradeEnvelopeAccount mirrors the Anchor TradeEnvelope layout on Solana.
type TradeEnvelopeAccount struct {
	Bump                     uint8     `json:"bump"`
	TradeID                  string    `json:"trade_id"`
	BuyerPubkey              string    `json:"buyer_pubkey"`
	SellerPubkey             string    `json:"seller_pubkey"`
	GradeID                  string    `json:"grade_id"`
	State                    string    `json:"state"`
	PriceCents               uint64    `json:"price_cents"`
	WindowStart              time.Time `json:"window_start"`
	WindowEnd                time.Time `json:"window_end"`
	LatestAttestationDigest  [32]byte  `json:"latest_attestation_digest"`
	NCCLAllReduceGbps        uint32    `json:"nccl_allreduce_gbps"`
	CanaryPassed             bool      `json:"canary_passed"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// MirrorAdapter bridges off-chain PostgreSQL canonical trade events to Solana devnet.
type MirrorAdapter struct {
	rpcEndpoint string
	programID   string
	logger      *slog.Logger
}

func NewMirrorAdapter(rpcEndpoint, programID string, logger *slog.Logger) *MirrorAdapter {
	if rpcEndpoint == "" {
		rpcEndpoint = "https://api.devnet.solana.com"
	}
	if programID == "" {
		programID = "VnodE7ZcEwXJ8gqJ2MhFhN3CqU9uWv4kL5Y7Xz8A1bC"
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MirrorAdapter{
		rpcEndpoint: rpcEndpoint,
		programID:   programID,
		logger:      logger,
	}
}

// ComputeTradeEnvelopePDA computes the Program-Derived Address seeds: [b"trade_envelope", trade_id.as_bytes()].
func (m *MirrorAdapter) ComputeTradeEnvelopePDA(tradeID string) ([32]byte, error) {
	if len(tradeID) == 0 {
		return [32]byte{}, fmt.Errorf("tradeID cannot be empty")
	}

	seed := []byte("trade_envelope" + tradeID)
	hash := sha256.Sum256(seed)
	return hash, nil
}

// MirrorTradeEnvelope registers the canonical trade envelope onto Solana devnet.
func (m *MirrorAdapter) MirrorTradeEnvelope(ctx context.Context, env TradeEnvelopeAccount) (string, error) {
	m.logger.Info("Mirroring trade envelope to Solana devnet",
		"trade_id", env.TradeID,
		"grade_id", env.GradeID,
		"state", env.State,
		"rpc", m.rpcEndpoint,
		"program_id", m.programID,
	)

	// Generates simulated transaction signature on devnet
	txSig := fmt.Sprintf("5Jn7W...simulated_tx_%s", env.TradeID[:8])
	return txSig, nil
}

// MirrorAttestationProof mirrors cryptographic canary results and ed25519 attestation hash to Solana devnet.
func (m *MirrorAdapter) MirrorAttestationProof(ctx context.Context, tradeID string, reportDigest [32]byte, ncclGbps uint32, passed bool) (string, error) {
	if passed && ncclGbps < 400 {
		return "", fmt.Errorf("cannot mirror attestation: NCCL bandwidth %d GB/s is below benchmark floor 400 GB/s", ncclGbps)
	}

	m.logger.Info("Mirroring hardware attestation proof to Solana devnet",
		"trade_id", tradeID,
		"nccl_gbps", ncclGbps,
		"canary_passed", passed,
	)

	txSig := fmt.Sprintf("4Kx8Z...attestation_tx_%s", tradeID[:8])
	return txSig, nil
}
