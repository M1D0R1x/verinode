package contract_test

import (
	"errors"
	"testing"

	"github.com/M1D0R1x/verinode/services/internal/contract"
)

func TestStateTransitions_TableDriven(t *testing.T) {
	t.Parallel()

	defaultEvt := contract.TransitionEvent{
		Actor:          "usr_buyer_123",
		Reason:         "valid business operation",
		IdempotencyKey: "idemp_test_key_001",
	}

	tests := []struct {
		name        string
		from        contract.State
		to          contract.State
		evt         contract.TransitionEvent
		wantErr     bool
		expectedErr error
	}{
		// Happy Path Transitions
		{
			name:    "Happy path: DraftRFQ to Open",
			from:    contract.StateDraftRFQ,
			to:      contract.StateOpen,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Open to Quoted",
			from:    contract.StateOpen,
			to:      contract.StateQuoted,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Quoted to Accepted",
			from:    contract.StateQuoted,
			to:      contract.StateAccepted,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Accepted to ContractPending",
			from:    contract.StateAccepted,
			to:      contract.StateContractPending,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: ContractPending to FundedSecured",
			from:    contract.StateContractPending,
			to:      contract.StateFundedSecured,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: FundedSecured to Scheduled",
			from:    contract.StateFundedSecured,
			to:      contract.StateScheduled,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Scheduled to DeliveryTest",
			from:    contract.StateScheduled,
			to:      contract.StateDeliveryTest,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: DeliveryTest to Live",
			from:    contract.StateDeliveryTest,
			to:      contract.StateLive,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Live to Completed",
			from:    contract.StateLive,
			to:      contract.StateCompleted,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Happy path: Completed to Settled",
			from:    contract.StateCompleted,
			to:      contract.StateSettled,
			evt:     defaultEvt,
			wantErr: false,
		},

		// Exception Paths: Cancellation before execution
		{
			name:    "Exception: Open to Cancelled",
			from:    contract.StateOpen,
			to:      contract.StateCancelled,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Exception: Quoted to Cancelled",
			from:    contract.StateQuoted,
			to:      contract.StateCancelled,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Exception: ContractPending to Cancelled",
			from:    contract.StateContractPending,
			to:      contract.StateCancelled,
			evt:     defaultEvt,
			wantErr: false,
		},

		// Exception Paths: Delivery Failure, Cure & Substitution
		{
			name:    "Remedy path: DeliveryTest to FailedDelivery",
			from:    contract.StateDeliveryTest,
			to:      contract.StateFailedDelivery,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Remedy path: FailedDelivery to Cure",
			from:    contract.StateFailedDelivery,
			to:      contract.StateCure,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Remedy path: Cure to Substituted",
			from:    contract.StateCure,
			to:      contract.StateSubstituted,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Remedy path: Substituted to DeliveryTest",
			from:    contract.StateSubstituted,
			to:      contract.StateDeliveryTest,
			evt:     defaultEvt,
			wantErr: false,
		},

		// Exception Paths: Claims & Disputes
		{
			name:    "Claims path: Live to ClaimOpen",
			from:    contract.StateLive,
			to:      contract.StateClaimOpen,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Claims path: ClaimOpen to Disputed",
			from:    contract.StateClaimOpen,
			to:      contract.StateDisputed,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Claims path: Disputed to Settled",
			from:    contract.StateDisputed,
			to:      contract.StateSettled,
			evt:     defaultEvt,
			wantErr: false,
		},
		{
			name:    "Claims path: Disputed to Terminated",
			from:    contract.StateDisputed,
			to:      contract.StateTerminated,
			evt:     defaultEvt,
			wantErr: false,
		},

		// Illegal Transitions (Must Fail)
		{
			name:        "Illegal: Settled to Open (Terminal state reactivation rejected)",
			from:        contract.StateSettled,
			to:          contract.StateOpen,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrTerminalState,
		},
		{
			name:        "Illegal: Cancelled to Live (Terminal state reactivation rejected)",
			from:        contract.StateCancelled,
			to:          contract.StateLive,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrTerminalState,
		},
		{
			name:        "Illegal: Terminated to Scheduled (Terminal state reactivation rejected)",
			from:        contract.StateTerminated,
			to:          contract.StateScheduled,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrTerminalState,
		},
		{
			name:        "Illegal: DraftRFQ straight to Live (Skipping essential stages rejected)",
			from:        contract.StateDraftRFQ,
			to:          contract.StateLive,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrIllegalTransition,
		},
		{
			name:        "Illegal: Live to DraftRFQ (Backwards regression rejected)",
			from:        contract.StateLive,
			to:          contract.StateDraftRFQ,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrIllegalTransition,
		},
		{
			name:        "Illegal: ContractPending to Live (Skipping funding rejected)",
			from:        contract.StateContractPending,
			to:          contract.StateLive,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrIllegalTransition,
		},
		{
			name:        "Illegal: Transition to identical state rejected",
			from:        contract.StateLive,
			to:          contract.StateLive,
			evt:         defaultEvt,
			wantErr:     true,
			expectedErr: contract.ErrSameState,
		},
		{
			name: "Validation error: Missing actor rejected",
			from: contract.StateDraftRFQ,
			to:   contract.StateOpen,
			evt: contract.TransitionEvent{
				Actor:          "",
				Reason:         "Valid reason",
				IdempotencyKey: "key_1",
			},
			wantErr:     true,
			expectedErr: contract.ErrMissingActor,
		},
		{
			name: "Validation error: Missing reason rejected",
			from: contract.StateDraftRFQ,
			to:   contract.StateOpen,
			evt: contract.TransitionEvent{
				Actor:          "usr_123",
				Reason:         "",
				IdempotencyKey: "key_1",
			},
			wantErr:     true,
			expectedErr: contract.ErrMissingReason,
		},
		{
			name: "Validation error: Missing idempotency key rejected",
			from: contract.StateDraftRFQ,
			to:   contract.StateOpen,
			evt: contract.TransitionEvent{
				Actor:          "usr_123",
				Reason:         "Valid reason",
				IdempotencyKey: "",
			},
			wantErr:     true,
			expectedErr: contract.ErrMissingIdempotency,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &contract.Contract{
				ID:    "trade_h100_test_uuid",
				State: tt.from,
			}

			event, err := contract.Transition(c, tt.to, tt.evt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Transition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Transition() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				return
			}

			if !tt.wantErr {
				if c.State != tt.to {
					t.Errorf("contract state = %v, want %v", c.State, tt.to)
				}
				if event == nil {
					t.Fatal("expected non-nil event on successful transition")
				}
				if event.PriorState != tt.from {
					t.Errorf("event.PriorState = %v, want %v", event.PriorState, tt.from)
				}
				if event.NewState != tt.to {
					t.Errorf("event.NewState = %v, want %v", event.NewState, tt.to)
				}
				if event.Actor != tt.evt.Actor {
					t.Errorf("event.Actor = %q, want %q", event.Actor, tt.evt.Actor)
				}
				if event.IdempotencyKey != tt.evt.IdempotencyKey {
					t.Errorf("event.IdempotencyKey = %q, want %q", event.IdempotencyKey, tt.evt.IdempotencyKey)
				}
			}
		})
	}
}

func TestCompleteHappyPath_LifecycleExecution(t *testing.T) {
	t.Parallel()

	c := &contract.Contract{
		ID:    "trade_lifecycle_complete_001",
		State: contract.StateDraftRFQ,
	}

	path := []contract.State{
		contract.StateOpen,
		contract.StateQuoted,
		contract.StateAccepted,
		contract.StateContractPending,
		contract.StateFundedSecured,
		contract.StateScheduled,
		contract.StateDeliveryTest,
		contract.StateLive,
		contract.StateCompleted,
		contract.StateSettled,
	}

	for i, step := range path {
		evt := contract.TransitionEvent{
			Actor:          "sys_orchestrator",
			Reason:         "stage completion",
			IdempotencyKey: "step_" + string(step),
		}

		event, err := contract.Transition(c, step, evt)
		if err != nil {
			t.Fatalf("Step %d (%s -> %s) failed unexpectedly: %v", i, c.State, step, err)
		}
		if c.State != step {
			t.Errorf("Contract state is %s, want %s", c.State, step)
		}
		if event.NewState != step {
			t.Errorf("Event state is %s, want %s", event.NewState, step)
		}
	}

	// Verify terminal state enforcement
	if !c.State.IsTerminal() {
		t.Errorf("Final state %s should be terminal", c.State)
	}

	_, err := contract.Transition(c, contract.StateOpen, contract.TransitionEvent{
		Actor:          "bad_actor",
		Reason:         "try to reopen",
		IdempotencyKey: "reopen_attempt",
	})
	if !errors.Is(err, contract.ErrTerminalState) {
		t.Fatalf("Reopening settled contract should yield ErrTerminalState, got: %v", err)
	}
}
