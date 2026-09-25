package contract

import "fmt"

// State represents the discrete lifecycle states of a Verinode capacity reservation contract.
type State string

const (
	// Happy path progression
	StateDraftRFQ        State = "draft_rfq"
	StateOpen            State = "open"
	StateQuoted          State = "quoted"
	StateAccepted        State = "accepted"
	StateContractPending State = "contract_pending"
	StateFundedSecured   State = "funded_secured"
	StateScheduled       State = "scheduled"
	StateDeliveryTest    State = "delivery_test"
	StateLive            State = "live"
	StateCompleted       State = "completed"
	StateSettled         State = "settled"

	// Exception & remedy states
	StateCancelled      State = "cancelled"
	StateFailedDelivery State = "failed_delivery"
	StateCure           State = "cure"
	StateSubstituted    State = "substituted"
	StateClaimOpen      State = "claim_open"
	StateDisputed       State = "disputed"
	StateTerminated     State = "terminated"
)

var validStates = map[State]bool{
	StateDraftRFQ:        true,
	StateOpen:            true,
	StateQuoted:          true,
	StateAccepted:        true,
	StateContractPending: true,
	StateFundedSecured:   true,
	StateScheduled:       true,
	StateDeliveryTest:    true,
	StateLive:            true,
	StateCompleted:       true,
	StateSettled:         true,
	StateCancelled:      true,
	StateFailedDelivery: true,
	StateCure:           true,
	StateSubstituted:    true,
	StateClaimOpen:      true,
	StateDisputed:       true,
	StateTerminated:     true,
}

var terminalStates = map[State]bool{
	StateSettled:    true,
	StateCancelled:  true,
	StateTerminated: true,
}

// IsValid reports whether s is a recognized contract state.
func (s State) IsValid() bool {
	return validStates[s]
}

// IsTerminal reports whether s is an immutable terminal state from which no further transitions are allowed.
func (s State) IsTerminal() bool {
	return terminalStates[s]
}

func (s State) String() string {
	return string(s)
}

// ParseState validates and converts a string into a State.
func ParseState(s string) (State, error) {
	st := State(s)
	if !st.IsValid() {
		return "", fmt.Errorf("invalid contract state: %q", s)
	}
	return st, nil
}
