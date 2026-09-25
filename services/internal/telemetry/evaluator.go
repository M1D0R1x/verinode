package telemetry

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrStaleReport = errors.New("telemetry: hardware report timestamp is older than 5 minutes")
	ErrFutureDate  = errors.New("telemetry: hardware report timestamp is in the future")
)

type ActionRecommendation string

const (
	ActionNone               ActionRecommendation = "none"
	ActionMarkDeliveryPassed ActionRecommendation = "mark_delivery_passed"
	ActionTriggerCure        ActionRecommendation = "trigger_cure"
	ActionOpenSLAClaim       ActionRecommendation = "open_sla_claim"
)

type EvaluationOutcome struct {
	Compliant            bool                 `json:"compliant"`
	Reason               string               `json:"reason"`
	CanaryPassed         bool                 `json:"canary_passed"`
	NCCLBandwidthGBPS    float64              `json:"nccl_bandwidth_gbps"`
	ECCErrors            int                  `json:"ecc_errors"`
	ActionRecommended    ActionRecommendation `json:"action_recommended"`
	EvaluatedAt          time.Time            `json:"evaluated_at"`
}

type CanaryEvaluator struct {
	pool *pgxpool.Pool
}

func NewCanaryEvaluator(pool *pgxpool.Pool) *CanaryEvaluator {
	return &CanaryEvaluator{pool: pool}
}

// Evaluate performs automated verification of a signed hardware attestation and canary result
// against benchmark specifications (e.g. 8x H100 SXM, 400 GB/s floor, 0 unrecovered ECC errors).
func (e *CanaryEvaluator) Evaluate(
	pubKey ed25519.PublicKey,
	report HardwareReport,
	reportSig []byte,
	canary CanaryResult,
	canarySig []byte,
	currentState string,
) (*EvaluationOutcome, error) {
	now := time.Now().UTC()

	// 1. Freshness check (within 5 minutes)
	if report.Timestamp.Before(now.Add(-5 * time.Minute)) {
		return nil, ErrStaleReport
	}
	if report.Timestamp.After(now.Add(1 * time.Minute)) {
		return nil, ErrFutureDate
	}

	// 2. Cryptographic signature check
	if err := VerifyHardwareAttestation(pubKey, report, reportSig); err != nil {
		return nil, fmt.Errorf("hardware attestation signature verification failed: %w", err)
	}
	if err := VerifyCanaryResult(pubKey, canary, canarySig); err != nil {
		return nil, fmt.Errorf("canary benchmark signature verification failed: %w", err)
	}

	outcome := &EvaluationOutcome{
		Compliant:         true,
		CanaryPassed:      canary.Passed,
		NCCLBandwidthGBPS: canary.NCCLAllReduceGBPS,
		ECCErrors:         report.ECCErrorsUnrecovered,
		EvaluatedAt:       now,
	}

	// 3. Topology & Hardware Validation (8x H100 SXM, 640GB VRAM, NVLink mesh)
	if err := ValidateAgainstGrade(report, 8, 640, "NVLink"); err != nil {
		outcome.Compliant = false
		outcome.Reason = fmt.Sprintf("Hardware specification breach: %v", err)
	} else if err := ValidateCanaryAgainstFloor(canary, 400.0); err != nil {
		outcome.Compliant = false
		outcome.Reason = fmt.Sprintf("Canary benchmark failed benchmark floor: %v", err)
	} else if report.ECCErrorsUnrecovered > 0 {
		outcome.Compliant = false
		outcome.Reason = fmt.Sprintf("Unrecovered ECC memory errors detected: %d", report.ECCErrorsUnrecovered)
	}

	// 4. Determine state-machine action recommendations per Phase 1.8 specifications
	if outcome.Compliant {
		if currentState == "delivery_test" {
			outcome.ActionRecommended = ActionMarkDeliveryPassed
			outcome.Reason = "Hardware and NCCL canary passed benchmark floors; ready for LIVE transition"
		} else {
			outcome.ActionRecommended = ActionNone
			outcome.Reason = "Routine telemetry heartbeat healthy; Invariant 4 satisfied"
		}
	} else {
		if currentState == "delivery_test" {
			outcome.ActionRecommended = ActionTriggerCure
		} else if currentState == "live" {
			outcome.ActionRecommended = ActionOpenSLAClaim
		} else {
			outcome.ActionRecommended = ActionNone
		}
	}

	return outcome, nil
}

// RecordDeliveryEvent persists the evaluation outcome into delivery_events table
func (e *CanaryEvaluator) RecordDeliveryEvent(
	ctx context.Context,
	contractID string,
	eventType string,
	evidenceRefs []string,
) error {
	if e.pool == nil {
		return nil // skip if DB is unconfigured
	}

	evidenceJSON, err := json.Marshal(evidenceRefs)
	if err != nil {
		evidenceJSON = []byte("[]")
	}

	query := `
		INSERT INTO delivery_events (contract_id, event_type, evidence_refs)
		VALUES ($1, $2, $3);
	`
	_, err = e.pool.Exec(ctx, query, contractID, eventType, evidenceJSON)
	if err != nil {
		return fmt.Errorf("persisting delivery event: %w", err)
	}

	return nil
}
