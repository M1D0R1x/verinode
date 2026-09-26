package workflow

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
)

func TestTradeLifecycleWorkflow(t *testing.T) {
	logger := slog.Default()
	activities := NewActivities(logger)
	orchestrator := NewOrchestrator(activities)
	ctx := context.Background()

	now := time.Now().UTC()
	baseInput := TradeWorkflowInput{
		TradeID:           "c37610f3-4aea-414f-868c-5937e97e822d",
		BuyerID:           "12cd814c-00d1-4c2f-b769-fa8854ce6bef",
		SellerID:          "91cb4f33-99bd-4597-9843-c22885e00862",
		GradeID:           "H100-SXM-8XNV",
		WindowStart:       now.Add(24 * time.Hour),
		WindowEnd:         now.Add(192 * time.Hour),
		EscrowAmountCents: 2956800, // $29,568.00
		Currency:          "USD",
		CanaryWindowLead:  24 * time.Hour,
	}

	t.Run("Happy path: compliant canary (428 GB/s) proceeds to Settled with balanced escrow payout", func(t *testing.T) {
		res, err := orchestrator.ExecuteTradeLifecycle(ctx, baseInput, 428.4)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if !res.CanaryPassed {
			t.Errorf("Expected canary to pass for 428.4 GB/s")
		}
		if res.FinalState != contract.StateSettled {
			t.Errorf("Expected final state %s, got %s", contract.StateSettled, res.FinalState)
		}
		if len(res.SettlementTxID) == 0 {
			t.Errorf("Expected valid settlement transaction ID")
		}
	})

	t.Run("Failure path: degraded canary (320 GB/s) triggers 100% escrow refund and cancels contract", func(t *testing.T) {
		res, err := orchestrator.ExecuteTradeLifecycle(ctx, baseInput, 320.0)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if res.CanaryPassed {
			t.Errorf("Expected canary to fail for 320.0 GB/s (floor is 400 GB/s)")
		}
		if res.FinalState != contract.StateCancelled {
			t.Errorf("Expected final state %s, got %s", contract.StateCancelled, res.FinalState)
		}
		if len(res.RefundTxID) == 0 {
			t.Errorf("Expected valid refund transaction ID for buyer")
		}
		if len(res.FailureReason) == 0 {
			t.Errorf("Expected descriptive failure reason")
		}
	})

	t.Run("Invalid funding: non-positive escrow amount cancels contract before canary", func(t *testing.T) {
		invalidInput := baseInput
		invalidInput.EscrowAmountCents = 0

		res, err := orchestrator.ExecuteTradeLifecycle(ctx, invalidInput, 428.4)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if res.FinalState != contract.StateCancelled {
			t.Errorf("Expected cancellation on unbacked escrow, got %s", res.FinalState)
		}
	})
}
