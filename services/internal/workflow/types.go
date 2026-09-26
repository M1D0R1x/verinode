package workflow

import (
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
)

// TaskQueueName defines the durable worker queue for Verinode trade lifecycle workflows.
const TaskQueueName = "verinode-trade-execution-queue"

// TradeWorkflowInput encapsulates all parameters required to drive a physical forward reservation.
type TradeWorkflowInput struct {
	TradeID           string    `json:"trade_id"`
	BuyerID           string    `json:"buyer_id"`
	SellerID          string    `json:"seller_id"`
	GradeID           string    `json:"grade_id"`
	WindowStart       time.Time `json:"window_start"`
	WindowEnd         time.Time `json:"window_end"`
	EscrowAmountCents int64     `json:"escrow_amount_cents"`
	Currency          string    `json:"currency"`
	CanaryWindowLead  time.Duration `json:"canary_window_lead"` // e.g. 24h before window start
}

// TradeWorkflowResult captures the final disposition and settlement state of the workflow.
type TradeWorkflowResult struct {
	TradeID         string         `json:"trade_id"`
	FinalState      contract.State `json:"final_state"`
	CanaryPassed    bool           `json:"canary_passed"`
	MeasuredGbps    float64        `json:"measured_gbps"`
	SettlementTxID  string         `json:"settlement_tx_id,omitempty"`
	RefundTxID      string         `json:"refund_tx_id,omitempty"`
	CompletedAt     time.Time      `json:"completed_at"`
	FailureReason   string         `json:"failure_reason,omitempty"`
}

// CanaryEvaluationResult returns the result of the host canary evaluation activity.
type CanaryEvaluationResult struct {
	Passed          bool    `json:"passed"`
	AllReduceGbps   float64 `json:"allreduce_gbps"`
	HealthyGPUs     int     `json:"healthy_gpus"`
	ECCErrors       int     `json:"ecc_errors"`
	ReportDigestHex string  `json:"report_digest_hex"`
	SignatureValid  bool    `json:"signature_valid"`
	Recommendation  string  `json:"recommendation"`
}
