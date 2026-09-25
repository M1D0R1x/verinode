package contract

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidState       = errors.New("contract: invalid state")
	ErrTerminalState      = errors.New("contract: cannot transition from terminal state")
	ErrSameState          = errors.New("contract: cannot transition to identical state")
	ErrIllegalTransition  = errors.New("contract: illegal state transition")
	ErrMissingActor       = errors.New("contract: transition event missing actor")
	ErrMissingReason      = errors.New("contract: transition event missing reason")
	ErrMissingIdempotency = errors.New("contract: transition event missing idempotency key")
)

// AllowedTransitions maps each state to its permissible successor states.
// Defined strictly according to docs/01-architecture-and-getting-started.md §1.5 and docs/03-api-specification.md §5.
var AllowedTransitions = map[State]map[State]bool{
	StateDraftRFQ: {
		StateOpen:      true,
		StateCancelled: true,
	},
	StateOpen: {
		StateQuoted:    true,
		StateCancelled: true,
	},
	StateQuoted: {
		StateAccepted:  true,
		StateCancelled: true,
	},
	StateAccepted: {
		StateContractPending: true,
		StateCancelled:       true,
	},
	StateContractPending: {
		StateFundedSecured: true,
		StateCancelled:     true,
	},
	StateFundedSecured: {
		StateScheduled: true,
		StateCancelled: true,
	},
	StateScheduled: {
		StateDeliveryTest: true,
		StateCancelled:    true,
	},
	StateDeliveryTest: {
		StateLive:           true,
		StateFailedDelivery: true,
	},
	StateLive: {
		StateCompleted:  true,
		StateClaimOpen:  true,
		StateTerminated: true,
	},
	StateCompleted: {
		StateSettled:   true,
		StateClaimOpen: true,
	},
	StateFailedDelivery: {
		StateCure:       true,
		StateTerminated: true,
	},
	StateCure: {
		StateSubstituted:  true,
		StateDeliveryTest: true,
		StateTerminated:   true,
	},
	StateSubstituted: {
		StateDeliveryTest: true,
		StateTerminated:   true,
	},
	StateClaimOpen: {
		StateDisputed:  true,
		StateLive:      true, // Claim resolved, capacity resumed
		StateCompleted: true, // Claim resolved, delivery completed
		StateSettled:   true, // Claim settled with SLA credit
		StateTerminated: true,
	},
	StateDisputed: {
		StateSettled:    true, // Arbitrated with settlement adjustments
		StateTerminated: true, // Arbitrated contract termination
	},
	// Terminal states have no successor states:
	StateSettled:    {},
	StateCancelled:  {},
	StateTerminated: {},
}

// TransitionEvent captures the audited, immutable metadata for a state change.
type TransitionEvent struct {
	Actor                 string                 `json:"actor"`
	Reason                string                 `json:"reason"`
	IdempotencyKey        string                 `json:"idempotency_key"`
	EvidenceHash          string                 `json:"evidence_hash,omitempty"`
	AuthorizationDecision map[string]interface{} `json:"authorization_decision,omitempty"`
}

// Contract represents the current state of a capacity reservation.
type Contract struct {
	ID        string    `json:"id"` // canonical trade_id
	State     State     `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContractEvent represents the recorded audit log entry for a transition.
type ContractEvent struct {
	ContractID            string                 `json:"contract_id"`
	PriorState            State                  `json:"prior_state"`
	NewState              State                  `json:"new_state"`
	Actor                 string                 `json:"actor"`
	Reason                string                 `json:"reason"`
	IdempotencyKey        string                 `json:"idempotency_key"`
	EvidenceHash          string                 `json:"evidence_hash,omitempty"`
	AuthorizationDecision map[string]interface{} `json:"authorization_decision,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
}

// CanTransition checks whether moving from current to next is legally permissible.
func CanTransition(current, next State) bool {
	if !current.IsValid() || !next.IsValid() {
		return false
	}
	if current.IsTerminal() || current == next {
		return false
	}
	successors, ok := AllowedTransitions[current]
	if !ok {
		return false
	}
	return successors[next]
}

// ValidateTransition returns a descriptive error if the transition is invalid.
func ValidateTransition(current, next State, evt TransitionEvent) error {
	if !current.IsValid() {
		return fmt.Errorf("%w: prior state %q is invalid", ErrInvalidState, current)
	}
	if !next.IsValid() {
		return fmt.Errorf("%w: target state %q is invalid", ErrInvalidState, next)
	}
	if current.IsTerminal() {
		return fmt.Errorf("%w: contract is already in terminal state %q", ErrTerminalState, current)
	}
	if current == next {
		return fmt.Errorf("%w: state is already %q", ErrSameState, current)
	}
	if !CanTransition(current, next) {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrIllegalTransition, current, next)
	}
	if evt.Actor == "" {
		return ErrMissingActor
	}
	if evt.Reason == "" {
		return ErrMissingReason
	}
	if evt.IdempotencyKey == "" {
		return ErrMissingIdempotency
	}
	return nil
}

// Transition advances the contract's state and produces an immutable audit event.
func Transition(c *Contract, next State, evt TransitionEvent) (*ContractEvent, error) {
	if c == nil {
		return nil, errors.New("contract: nil contract pointer")
	}

	if err := ValidateTransition(c.State, next, evt); err != nil {
		return nil, err
	}

	prior := c.State
	now := time.Now().UTC()

	c.State = next
	c.UpdatedAt = now

	return &ContractEvent{
		ContractID:            c.ID,
		PriorState:            prior,
		NewState:              next,
		Actor:                 evt.Actor,
		Reason:                evt.Reason,
		IdempotencyKey:        evt.IdempotencyKey,
		EvidenceHash:          evt.EvidenceHash,
		AuthorizationDecision: evt.AuthorizationDecision,
		CreatedAt:             now,
	}, nil
}
