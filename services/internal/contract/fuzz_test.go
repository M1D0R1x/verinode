package contract

import (
	"testing"
)

func FuzzStateTransitionLegality(f *testing.F) {
	// Seed corpus with valid and invalid state pairs
	f.Add("draft_rfq", "open")
	f.Add("open", "quoted")
	f.Add("quoted", "accepted")
	f.Add("accepted", "contract_pending")
	f.Add("contract_pending", "funded_secured")
	f.Add("funded_secured", "scheduled")
	f.Add("scheduled", "delivery_test")
	f.Add("delivery_test", "live")
	f.Add("live", "completed")
	f.Add("completed", "settled")
	f.Add("settled", "open")       // Illegal: terminal state reactivation
	f.Add("live", "settled")        // Illegal: skip completed
	f.Add("random_state", "live")   // Illegal: unknown state
	f.Add("", "")                   // Illegal: empty state

	evt := TransitionEvent{
		Actor:          "fuzz-runner",
		Reason:         "Automated state machine fuzzing",
		IdempotencyKey: "fuzz-key-001",
	}

	f.Fuzz(func(t *testing.T, fromStr, toStr string) {
		from := State(fromStr)
		to := State(toStr)

		err := ValidateTransition(from, to, evt)
		can := CanTransition(from, to)

		// Property 1: If either state is not valid, it MUST fail
		if !from.IsValid() || !to.IsValid() {
			if err == nil || can {
				t.Fatalf("Expected transition between invalid states (%q -> %q) to fail", from, to)
			}
			return
		}

		// Property 2: Self-transitions must ALWAYS fail
		if from == to {
			if err == nil || can {
				t.Fatalf("Expected self-transition (%q -> %q) to fail", from, to)
			}
			return
		}

		// Property 3: Terminal states must NEVER transition anywhere
		if from.IsTerminal() {
			if err == nil || can {
				t.Fatalf("Expected transition from terminal state %q to %q to fail", from, to)
			}
			return
		}

		// Property 4: ValidateTransition error == nil MUST strictly agree with CanTransition
		if (err == nil) != can {
			t.Fatalf("Disagreement between ValidateTransition (err=%v) and CanTransition (%v) for %q -> %q", err, can, from, to)
		}

		// Property 5: If valid, the transition MUST be in AllowedTransitions
		if err == nil {
			targets, ok := AllowedTransitions[from]
			if !ok {
				t.Fatalf("Transition %q -> %q passed validation but has no AllowedTransitions map entry", from, to)
			}
			if !targets[to] {
				t.Fatalf("Transition %q -> %q passed validation but is not in targets map", from, to)
			}
		}
	})
}
