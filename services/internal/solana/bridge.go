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
//
// This adapter mirrors the canonical off-chain trade envelope and its cryptographic
// canary attestation onto Solana devnet by anchoring a compact memo containing the
// trade_id and a SHA-256 digest. When a funded signer is configured
// (SOLANA_SETTLEMENT_KEYPAIR) it submits a REAL devnet transaction and returns a
// publicly verifiable signature + Explorer URL; otherwise it returns a deterministic
// SIMULATED- proof so the product runs end-to-end without funded keys.

// TradeEnvelopeAccount mirrors the Anchor TradeEnvelope layout on Solana.
type TradeEnvelopeAccount struct {
	Bump                    uint8     `json:"bump"`
	TradeID                 string    `json:"trade_id"`
	BuyerPubkey             string    `json:"buyer_pubkey"`
	SellerPubkey            string    `json:"seller_pubkey"`
	GradeID                 string    `json:"grade_id"`
	State                   string    `json:"state"`
	PriceCents              uint64    `json:"price_cents"`
	WindowStart             time.Time `json:"window_start"`
	WindowEnd               time.Time `json:"window_end"`
	LatestAttestationDigest [32]byte  `json:"latest_attestation_digest"`
	NCCLAllReduceGbps       uint32    `json:"nccl_allreduce_gbps"`
	CanaryPassed            bool      `json:"canary_passed"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// MirrorResult captures the outcome of a mirror operation, including whether it
// was a real on-chain submission and where to verify it.
type MirrorResult struct {
	Signature   string `json:"signature"`
	ExplorerURL string `json:"explorer_url,omitempty"`
	Mode        string `json:"mode"` // "live-devnet" | "simulated"
	PDA         string `json:"pda,omitempty"`
	Memo        string `json:"memo"`
}

// MirrorAdapter bridges off-chain PostgreSQL canonical trade events to Solana devnet.
type MirrorAdapter struct {
	rpcEndpoint string
	programID   string
	logger      *slog.Logger
	client      *DevnetClient
}

// DefaultProgramID is the declared Anchor program address for the Verinode registry.
const DefaultProgramID = "VnodE7ZcEwXJ8gqJ2MhFhN3CqU9uWv4kL5Y7Xz8A1bC"

func NewMirrorAdapter(rpcEndpoint, programID string, logger *slog.Logger) *MirrorAdapter {
	if rpcEndpoint == "" {
		rpcEndpoint = "https://api.devnet.solana.com"
	}
	if programID == "" {
		programID = DefaultProgramID
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MirrorAdapter{
		rpcEndpoint: rpcEndpoint,
		programID:   programID,
		logger:      logger,
		client:      NewDevnetClient(rpcEndpoint, ""),
	}
}

// Mode reports whether the adapter will submit real devnet transactions.
func (m *MirrorAdapter) Mode() string { return m.client.Mode() }

// SignerPubkey returns the configured devnet signer address (or zero address).
func (m *MirrorAdapter) SignerPubkey() string { return m.client.Pubkey() }

// ComputeTradeEnvelopePDA computes the Program-Derived Address seeds: [b"trade_envelope", trade_id.as_bytes()].
// This mirrors the Anchor program's PDA derivation shape (SHA-256 over the seed set
// is used here as a deterministic, dependency-free stand-in for the on-chain
// find_program_address, which the deployed Anchor program computes canonically).
func (m *MirrorAdapter) ComputeTradeEnvelopePDA(tradeID string) ([32]byte, error) {
	if len(tradeID) == 0 {
		return [32]byte{}, fmt.Errorf("tradeID cannot be empty")
	}

	seed := []byte("trade_envelope" + tradeID)
	hash := sha256.Sum256(seed)
	return hash, nil
}

// MirrorTradeEnvelope anchors the canonical trade envelope onto Solana devnet and
// returns the transaction signature. Preserved signature for backward compatibility.
func (m *MirrorAdapter) MirrorTradeEnvelope(ctx context.Context, env TradeEnvelopeAccount) (string, error) {
	res, err := m.MirrorTradeEnvelopeResult(ctx, env)
	if err != nil {
		return "", err
	}
	return res.Signature, nil
}

// MirrorTradeEnvelopeResult anchors the trade state and returns rich proof metadata.
func (m *MirrorAdapter) MirrorTradeEnvelopeResult(ctx context.Context, env TradeEnvelopeAccount) (MirrorResult, error) {
	if env.TradeID == "" {
		return MirrorResult{}, fmt.Errorf("trade_id is required to mirror envelope")
	}

	pda, _ := m.ComputeTradeEnvelopePDA(env.TradeID)
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d", env.TradeID, env.GradeID, env.State, env.PriceCents)))
	memo := deterministicMemoPayload("state:"+env.State, env.TradeID, digest, 0)

	sig, err := m.client.AnchorMemo(ctx, memo)
	if err != nil {
		return MirrorResult{}, err
	}

	m.logger.Info("Mirrored trade envelope to Solana",
		"trade_id", env.TradeID,
		"grade_id", env.GradeID,
		"state", env.State,
		"mode", m.client.Mode(),
		"signature", sig,
	)

	return MirrorResult{
		Signature:   sig,
		ExplorerURL: ExplorerURL(sig, "devnet"),
		Mode:        m.client.Mode(),
		PDA:         base58Encode(pda[:]),
		Memo:        memo,
	}, nil
}

// MirrorAttestationProof mirrors cryptographic canary results and the ed25519
// attestation digest to Solana devnet. Preserved signature.
func (m *MirrorAdapter) MirrorAttestationProof(ctx context.Context, tradeID string, reportDigest [32]byte, ncclGbps uint32, passed bool) (string, error) {
	res, err := m.MirrorAttestationProofResult(ctx, tradeID, reportDigest, ncclGbps, passed)
	if err != nil {
		return "", err
	}
	return res.Signature, nil
}

// MirrorAttestationProofResult mirrors the canary attestation and returns rich proof metadata.
func (m *MirrorAdapter) MirrorAttestationProofResult(ctx context.Context, tradeID string, reportDigest [32]byte, ncclGbps uint32, passed bool) (MirrorResult, error) {
	if tradeID == "" {
		return MirrorResult{}, fmt.Errorf("trade_id is required to mirror attestation")
	}
	// Invariant: never anchor a "passed" proof that is below the benchmark floor.
	if passed && ncclGbps < 400 {
		return MirrorResult{}, fmt.Errorf("cannot mirror attestation: NCCL bandwidth %d GB/s is below benchmark floor 400 GB/s", ncclGbps)
	}

	memo := deterministicMemoPayload("canary", tradeID, reportDigest, ncclGbps)
	sig, err := m.client.AnchorMemo(ctx, memo)
	if err != nil {
		return MirrorResult{}, err
	}

	m.logger.Info("Mirrored hardware attestation proof to Solana",
		"trade_id", tradeID,
		"nccl_gbps", ncclGbps,
		"canary_passed", passed,
		"mode", m.client.Mode(),
		"signature", sig,
	)

	return MirrorResult{
		Signature:   sig,
		ExplorerURL: ExplorerURL(sig, "devnet"),
		Mode:        m.client.Mode(),
		Memo:        memo,
	}, nil
}
