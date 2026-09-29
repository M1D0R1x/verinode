// Package chain orchestrates the OPTIONAL, Phase-4 on-chain audit mirrors across
// Solana, Arbitrum and Hyperliquid. Per Invariant 6 these are read-only mirrors of
// the canonical off-chain PostgreSQL trade record — never a second source of truth,
// never in the RFQ→contract→funding→delivery→settlement critical path.
//
// The orchestrator anchors a canonical trade envelope + its cryptographic canary
// attestation on each configured rail and returns a single combined proof bundle
// that the API and UI surface. Each rail degrades independently to a labelled
// SIMULATED proof when no funded key is configured, so the product always runs.
package chain

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/arbitrum"
	"github.com/M1D0R1x/verinode/services/internal/hyperliquid"
	"github.com/M1D0R1x/verinode/services/internal/solana"
)

// TradeEnvelope is the chain-agnostic canonical trade record the orchestrator mirrors.
type TradeEnvelope struct {
	TradeID       string    `json:"trade_id"`
	BuyerID       string    `json:"buyer_id"`
	SellerID      string    `json:"seller_id"`
	GradeID       string    `json:"grade_id"`
	State         string    `json:"state"`
	PriceCents    uint64    `json:"price_cents"`
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
	NCCLGbps      uint32    `json:"nccl_gbps"`
	CanaryPassed  bool      `json:"canary_passed"`
	AttestDigest  [32]byte  `json:"-"`
}

// RailProof captures one chain's mirror result.
type RailProof struct {
	Chain       string `json:"chain"`
	Mode        string `json:"mode"`
	StateTx     string `json:"state_tx,omitempty"`
	AttestTx    string `json:"attestation_tx,omitempty"`
	ExplorerURL string `json:"explorer_url,omitempty"`
	Reference   string `json:"reference,omitempty"` // PDA / contract / oracle id
	Error       string `json:"error,omitempty"`
}

// ProofBundle is the combined multi-chain mirror result for one canonical trade.
type ProofBundle struct {
	TradeID   string      `json:"trade_id"`
	MirroredAt time.Time  `json:"mirrored_at"`
	Rails     []RailProof `json:"rails"`
	Note      string      `json:"note"`
}

// Orchestrator holds the per-chain adapters.
type Orchestrator struct {
	sol    *solana.MirrorAdapter
	arb    *arbitrum.ArbitrumAdapter
	hl     *hyperliquid.Adapter
	logger *slog.Logger
}

// Config configures the orchestrator's endpoints. Empty values fall back to
// public devnet/testnet defaults inside each adapter.
type Config struct {
	SolanaRPC       string
	SolanaProgramID string
	ArbitrumRPC     string
	ArbitrumAddr    string
	HyperliquidAPI  string
	HyperliquidOID  string
}

// NewOrchestrator builds an orchestrator from config.
func NewOrchestrator(cfg Config, logger *slog.Logger) *Orchestrator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Orchestrator{
		sol:    solana.NewMirrorAdapter(cfg.SolanaRPC, cfg.SolanaProgramID, logger),
		arb:    arbitrum.NewArbitrumAdapter(cfg.ArbitrumRPC, cfg.ArbitrumAddr, logger),
		hl:     hyperliquid.NewAdapter(cfg.HyperliquidAPI, cfg.HyperliquidOID, logger),
		logger: logger,
	}
}

// Modes reports the live/simulated mode of each rail (for status surfaces).
func (o *Orchestrator) Modes() map[string]string {
	return map[string]string{
		"solana":      o.sol.Mode(),
		"arbitrum":    o.arb.Mode(),
		"hyperliquid": o.hl.Mode(),
	}
}

// MirrorTrade anchors the canonical trade envelope and its canary attestation on
// Solana and Arbitrum, returning a combined proof bundle. Each rail is independent:
// one rail failing never aborts the others (Invariant 6 no-bridging: each rail
// reconciles independently against the one canonical trade_id).
func (o *Orchestrator) MirrorTrade(ctx context.Context, env TradeEnvelope) ProofBundle {
	bundle := ProofBundle{
		TradeID:    env.TradeID,
		MirroredAt: time.Now().UTC(),
		Note:       "Off-chain PostgreSQL remains authoritative (Invariant 6). These are read-only audit mirrors.",
	}

	// --- Solana rail ---
	solProof := RailProof{Chain: "solana", Mode: o.sol.Mode()}
	solEnv := solana.TradeEnvelopeAccount{
		TradeID:           env.TradeID,
		BuyerPubkey:       env.BuyerID,
		SellerPubkey:      env.SellerID,
		GradeID:           env.GradeID,
		State:             env.State,
		PriceCents:        env.PriceCents,
		WindowStart:       env.WindowStart,
		WindowEnd:         env.WindowEnd,
		NCCLAllReduceGbps: env.NCCLGbps,
		CanaryPassed:      env.CanaryPassed,
		UpdatedAt:         time.Now().UTC(),
	}
	if res, err := o.sol.MirrorTradeEnvelopeResult(ctx, solEnv); err != nil {
		solProof.Error = err.Error()
	} else {
		solProof.StateTx = res.Signature
		solProof.ExplorerURL = res.ExplorerURL
		solProof.Reference = res.PDA
	}
	if env.CanaryPassed {
		if ares, err := o.sol.MirrorAttestationProofResult(ctx, env.TradeID, env.AttestDigest, env.NCCLGbps, true); err != nil {
			if solProof.Error == "" {
				solProof.Error = err.Error()
			}
		} else {
			solProof.AttestTx = ares.Signature
			if solProof.ExplorerURL == "" {
				solProof.ExplorerURL = ares.ExplorerURL
			}
		}
	}
	bundle.Rails = append(bundle.Rails, solProof)

	// --- Arbitrum rail ---
	arbProof := RailProof{Chain: "arbitrum", Mode: o.arb.Mode()}
	tradeBytes, _ := o.arb.ComputeTradeBytes32(env.TradeID)
	arbEnv := arbitrum.TradeEnvelopeEVM{
		TradeID:         tradeBytes,
		BuyerAddress:    env.BuyerID,
		SellerAddress:   env.SellerID,
		EscrowAmountWei: centsToWei(env.PriceCents),
		WindowStart:     uint64(env.WindowStart.Unix()),
		WindowEnd:       uint64(env.WindowEnd.Unix()),
		State:           stateToEVM(env.State),
		NCCLGbps:        env.NCCLGbps,
		CanaryPassed:    env.CanaryPassed,
	}
	if res, err := o.arb.MirrorTradeResult(ctx, arbEnv); err != nil {
		arbProof.Error = err.Error()
	} else {
		arbProof.StateTx = res.TxHash
		arbProof.ExplorerURL = res.ExplorerURL
		arbProof.Reference = res.Contract
	}
	if env.CanaryPassed {
		if tx, err := o.arb.MirrorAttestationToArbitrum(ctx, tradeBytes, env.AttestDigest, env.NCCLGbps, true); err != nil {
			if arbProof.Error == "" {
				arbProof.Error = err.Error()
			}
		} else {
			arbProof.AttestTx = tx
		}
	}
	bundle.Rails = append(bundle.Rails, arbProof)

	return bundle
}

// Hedge computes a Hyperliquid delta-hedge advisory for a physical forward.
func (o *Orchestrator) Hedge(durationHours, gpuCount int, fixedRateHourly, indexMarkPrice float64) hyperliquid.HedgePositionQuote {
	return o.hl.CalculateHedgeQuote(durationHours, gpuCount, fixedRateHourly, indexMarkPrice)
}

// PublishOracle mirrors the signed index fix to the Hyperliquid HIP-3 oracle,
// enforcing Invariant 5 (insufficient_data / min contributors / positive mark).
func (o *Orchestrator) PublishOracle(ctx context.Context, u hyperliquid.HIP3OracleUpdate) (string, error) {
	return o.hl.PublishHIP3OracleFeed(ctx, u)
}

// centsToWei converts USD cents to a wei-scaled big.Int for the EVM escrow field
// (illustrative 1e12 scaling — the on-chain value is a mirror, never authoritative).
func centsToWei(cents uint64) *big.Int {
	w := new(big.Int).SetUint64(cents)
	return w.Mul(w, big.NewInt(10_000_000_000)) // cents * 1e10 => wei-ish reference
}

// stateToEVM maps the off-chain contract state string to the Solidity enum ordinal
// used by VerinodeRegistry.sol.
func stateToEVM(state string) uint8 {
	switch state {
	case "contract_pending":
		return 0
	case "funded_secured":
		return 1
	case "scheduled":
		return 2
	case "delivery_test":
		return 3
	case "live":
		return 4
	case "completed":
		return 5
	case "settled":
		return 6
	case "cancelled":
		return 7
	case "failed_delivery":
		return 8
	default:
		return 0
	}
}
