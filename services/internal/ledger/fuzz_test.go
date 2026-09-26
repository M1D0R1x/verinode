package ledger

import (
	"errors"
	"math"
	"testing"
	"time"
)

// FuzzValidateTransaction_BalancedInvariant performs property-based fuzzing to verify
// Invariant 3: SUM(amount_cents) == 0 grouped by transaction_id.
func FuzzValidateTransaction_BalancedInvariant(f *testing.F) {
	// Seed corpus with interesting boundary values
	f.Add(int64(100), int64(100), int64(0), true)
	f.Add(int64(2956800), int64(2956800), int64(50000), true)
	f.Add(int64(1), int64(2), int64(0), false)
	f.Add(int64(0), int64(0), int64(0), false)
	f.Add(int64(math.MaxInt32), int64(math.MaxInt32), int64(100), true)

	f.Fuzz(func(t *testing.T, amount1, amount2, fee int64, forceBalanced bool) {
		txID := "fuzz-tx-001"
		currency := "USD"
		now := time.Now().UTC()

		if forceBalanced {
			// Build an intentionally balanced transaction:
			if amount1 <= 0 {
				return
			}

			var entries []Entry
			if fee > 0 && fee < amount1 {
				// 3-legged settlement
				sellerAmt := amount1 - fee
				entries = []Entry{
					{TransactionID: txID, AccountID: "acct-escrow", AmountCents: -amount1, Currency: currency, CreatedAt: now},
					{TransactionID: txID, AccountID: "acct-seller", AmountCents: sellerAmt, Currency: currency, CreatedAt: now},
					{TransactionID: txID, AccountID: "acct-platform", AmountCents: fee, Currency: currency, CreatedAt: now},
				}
			} else {
				// 2-legged direct transfer
				entries = []Entry{
					{TransactionID: txID, AccountID: "acct-escrow", AmountCents: -amount1, Currency: currency, CreatedAt: now},
					{TransactionID: txID, AccountID: "acct-seller", AmountCents: amount1, Currency: currency, CreatedAt: now},
				}
			}

			err := ValidateTransaction(entries)
			if err != nil {
				t.Fatalf("Expected balanced transaction to pass, got: %v", err)
			}
		} else {
			// Build an intentionally unbalanced transaction where amount1 != amount2 and neither is 0
			if amount1 == 0 {
				amount1 = 1
			}
			if amount2 == 0 {
				amount2 = 2
			}
			if amount1 == amount2 {
				amount2 = amount1 + 1
			}

			entries := []Entry{
				{TransactionID: txID, AccountID: "acct-1", AmountCents: amount1, Currency: currency, CreatedAt: now},
				{TransactionID: txID, AccountID: "acct-2", AmountCents: -amount2, Currency: currency, CreatedAt: now},
			}

			err := ValidateTransaction(entries)
			if err == nil {
				t.Fatalf("Expected unbalanced transaction to fail (amount1=%d, amount2=%d), got nil", amount1, amount2)
			}
			if !errors.Is(err, ErrUnbalancedTransaction) {
				t.Fatalf("Expected ErrUnbalancedTransaction, got: %v", err)
			}
		}
	})
}

// FuzzBuyerDepositTransaction verifies that deposit transactions always satisfy Invariant 3.
func FuzzBuyerDepositTransaction(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(2956800))
	f.Add(int64(0))
	f.Add(int64(-100))
	f.Add(int64(math.MaxInt32))

	f.Fuzz(func(t *testing.T, amount int64) {
		entries, err := NewBuyerDepositTransaction(
			"tx-fuzz-deposit",
			"acct-escrow",
			"acct-buyer",
			"trade-fuzz-01",
			amount,
			"USD",
			"BANK-REF-FUZZ",
		)

		if amount <= 0 {
			if err == nil {
				t.Fatalf("Expected non-positive deposit amount %d to be rejected, got nil", amount)
			}
			return
		}

		if err != nil {
			t.Fatalf("Expected positive deposit %d to succeed, got: %v", amount, err)
		}

		// Verify invariant SUM == 0
		var sum int64
		for _, e := range entries {
			sum += e.AmountCents
		}
		if sum != 0 {
			t.Fatalf("Invariant 3 broken! Sum of entries is %d, expected 0", sum)
		}
	})
}
