package telemetry_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

func TestCanaryEvaluator_Scenarios(t *testing.T) {
	t.Parallel()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed generating key: %v", err)
	}

	evaluator := telemetry.NewCanaryEvaluator(nil)

	now := time.Now().UTC()

	validReport := telemetry.HardwareReport{
		AgentID:              "agent_001",
		BlockID:              "block_001",
		Timestamp:            now,
		GPUSKU:               "NVIDIA H100 SXM 80GB",
		GPUCount:             8,
		MemoryGB:             640,
		Topology:             "SXM5/HGX NVLink 4.0",
		DriverVersion:        "535.129.03",
		ECCErrorsUnrecovered: 0,
		SequenceNumber:       1,
	}
	reportPayload, _ := validReport.CanonicalPayload()
	validReportSig := ed25519.Sign(privKey, reportPayload)

	validCanary := telemetry.CanaryResult{
		AgentID:           "agent_001",
		BlockID:           "block_001",
		Timestamp:         now,
		TestName:          "nccl_allreduce_perf",
		Passed:            true,
		NCCLAllReduceGBPS: 428.5,
	}
	canaryPayload, _ := validCanary.CanonicalPayload()
	validCanarySig := ed25519.Sign(privKey, canaryPayload)

	t.Run("Compliant cluster in delivery_test recommends mark_delivery_passed", func(t *testing.T) {
		t.Parallel()

		outcome, err := evaluator.Evaluate(
			pubKey,
			validReport,
			validReportSig,
			validCanary,
			validCanarySig,
			"delivery_test",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !outcome.Compliant {
			t.Errorf("expected compliant outcome, got %v: %s", outcome.Compliant, outcome.Reason)
		}
		if outcome.ActionRecommended != telemetry.ActionMarkDeliveryPassed {
			t.Errorf("expected action mark_delivery_passed, got %s", outcome.ActionRecommended)
		}
	})

	t.Run("Canary below 400 GB/s in delivery_test recommends trigger_cure", func(t *testing.T) {
		t.Parallel()

		badCanary := validCanary
		badCanary.NCCLAllReduceGBPS = 345.0 // Below 400 GB/s floor
		badCanaryPayload, _ := badCanary.CanonicalPayload()
		badCanarySig := ed25519.Sign(privKey, badCanaryPayload)

		outcome, err := evaluator.Evaluate(
			pubKey,
			validReport,
			validReportSig,
			badCanary,
			badCanarySig,
			"delivery_test",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outcome.Compliant {
			t.Errorf("expected non-compliant outcome for 345 GB/s")
		}
		if outcome.ActionRecommended != telemetry.ActionTriggerCure {
			t.Errorf("expected action trigger_cure, got %s", outcome.ActionRecommended)
		}
	})

	t.Run("Unrecovered ECC memory error in live recommends open_sla_claim", func(t *testing.T) {
		t.Parallel()

		eccReport := validReport
		eccReport.ECCErrorsUnrecovered = 2
		eccPayload, _ := eccReport.CanonicalPayload()
		eccSig := ed25519.Sign(privKey, eccPayload)

		outcome, err := evaluator.Evaluate(
			pubKey,
			eccReport,
			eccSig,
			validCanary,
			validCanarySig,
			"live",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outcome.Compliant {
			t.Errorf("expected non-compliant outcome for ECC errors")
		}
		if outcome.ActionRecommended != telemetry.ActionOpenSLAClaim {
			t.Errorf("expected action open_sla_claim, got %s", outcome.ActionRecommended)
		}
	})

	t.Run("Stale telemetry report older than 5 minutes is rejected", func(t *testing.T) {
		t.Parallel()

		staleReport := validReport
		staleReport.Timestamp = now.Add(-10 * time.Minute)
		stalePayload, _ := staleReport.CanonicalPayload()
		staleSig := ed25519.Sign(privKey, stalePayload)

		_, err := evaluator.Evaluate(
			pubKey,
			staleReport,
			staleSig,
			validCanary,
			validCanarySig,
			"live",
		)
		if err == nil {
			t.Fatalf("expected error for stale report, but passed!")
		}
	})
}
