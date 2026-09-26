package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
)

// Orchestrator coordinates durable execution of a physical forward reservation contract.
type Orchestrator struct {
	activities *Activities
}

func NewOrchestrator(activities *Activities) *Orchestrator {
	return &Orchestrator{activities: activities}
}

// ExecuteTradeLifecycle runs the end-to-end durable state progression.
func (o *Orchestrator) ExecuteTradeLifecycle(ctx context.Context, input TradeWorkflowInput, simulatedNCCLGbps float64) (*TradeWorkflowResult, error) {
	result := &TradeWorkflowResult{
		TradeID:      input.TradeID,
		FinalState:   contract.StateContractPending,
		CanaryPassed: false,
		CompletedAt:  time.Now().UTC(),
	}

	// Step 1: Verify e-Signatures (ContractPending -> FundedSecured)
	_, err := o.activities.VerifySignaturesActivity(ctx, input)
	if err != nil {
		result.FailureReason = fmt.Sprintf("signature verification failed: %v", err)
		result.FinalState = contract.StateCancelled
		return result, nil
	}

	// Step 2: Verify Escrow Funding (FundedSecured -> Scheduled)
	funded, err := o.activities.VerifyEscrowFundingActivity(ctx, input)
	if err != nil || !funded {
		result.FailureReason = fmt.Sprintf("escrow funding verification failed: %v", err)
		result.FinalState = contract.StateCancelled
		return result, nil
	}
	result.FinalState = contract.StateScheduled

	// Step 3: Canary Evaluation Gate (Scheduled -> DeliveryTest -> Live or FailedDelivery)
	canary, err := o.activities.ExecuteCanaryEvaluationActivity(ctx, input, simulatedNCCLGbps)
	if err != nil {
		result.FailureReason = fmt.Sprintf("canary evaluation failed: %v", err)
		result.FinalState = contract.StateFailedDelivery
		return result, nil
	}

	result.MeasuredGbps = canary.AllReduceGbps
	result.CanaryPassed = canary.Passed

	if !canary.Passed {
		// Hardware failed benchmark floor (e.g. < 400 GB/s NCCL AllReduce)
		// Process 100% Escrow refund to buyer
		refundTx, err := o.activities.ExecuteEscrowRefundActivity(ctx, input, "CANARY_FLOOR_FAILED")
		if err != nil {
			return nil, fmt.Errorf("fatal: failed to process buyer refund: %w", err)
		}
		result.RefundTxID = refundTx
		result.FinalState = contract.StateCancelled
		result.FailureReason = fmt.Sprintf("hardware canary below benchmark floor (%.1f GB/s < 400.0 GB/s); 100%% escrow refunded", canary.AllReduceGbps)
		return result, nil
	}

	// Canary Passed: Transition to Live Delivery
	result.FinalState = contract.StateLive

	// Step 4: Window Delivery Completion & Escrow Payout (Live -> Completed -> Settled)
	settleTx, err := o.activities.ExecuteEscrowSettlementActivity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("fatal: failed to settle escrow: %w", err)
	}

	result.SettlementTxID = settleTx
	result.FinalState = contract.StateSettled
	result.CompletedAt = time.Now().UTC()

	return result, nil
}
