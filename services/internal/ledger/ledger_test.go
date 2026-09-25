package ledger_test

import (
	"errors"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/ledger"
)

func TestValidateTransaction_TableDriven(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name        string
		entries     []ledger.Entry
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Valid: standard 2-legged balanced transaction",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_001",
					AccountID:     "acc_escrow",
					AmountCents:   500000, // $5,000.00
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_001",
					AccountID:     "acc_buyer_dep",
					AmountCents:   -500000,
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid: 3-legged settlement with fee",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_002",
					AccountID:     "acc_escrow",
					AmountCents:   -1000000, // -$10,000.00
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_002",
					AccountID:     "acc_seller",
					AmountCents:   950000, // +$9,500.00
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_002",
					AccountID:     "acc_verinode_fee",
					AmountCents:   50000, // +$500.00
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid: Unbalanced transaction (SUM != 0) must be rejected",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_unbal",
					AccountID:     "acc_escrow",
					AmountCents:   500000,
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_unbal",
					AccountID:     "acc_buyer",
					AmountCents:   -499999, // 1 cent discrepancy!
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr:     true,
			expectedErr: ledger.ErrUnbalancedTransaction,
		},
		{
			name: "Invalid: Single entry transaction must be rejected",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_single",
					AccountID:     "acc_escrow",
					AmountCents:   0,
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr:     true,
			expectedErr: ledger.ErrEmptyTransaction,
		},
		{
			name: "Invalid: Mixed currencies must be rejected",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_cur",
					AccountID:     "acc_1",
					AmountCents:   1000,
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_cur",
					AccountID:     "acc_2",
					AmountCents:   -1000,
					Currency:      "EUR", // Mismatch
					CreatedAt:     now,
				},
			},
			wantErr:     true,
			expectedErr: ledger.ErrMixedCurrencies,
		},
		{
			name: "Invalid: Zero amount entry must be rejected",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_zero",
					AccountID:     "acc_1",
					AmountCents:   0,
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_zero",
					AccountID:     "acc_2",
					AmountCents:   0,
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr:     true,
			expectedErr: ledger.ErrZeroAmountEntry,
		},
		{
			name: "Invalid: Missing account ID must be rejected",
			entries: []ledger.Entry{
				{
					TransactionID: "tx_no_acc",
					AccountID:     "",
					AmountCents:   500,
					Currency:      "USD",
					CreatedAt:     now,
				},
				{
					TransactionID: "tx_no_acc",
					AccountID:     "acc_2",
					AmountCents:   -500,
					Currency:      "USD",
					CreatedAt:     now,
				},
			},
			wantErr:     true,
			expectedErr: ledger.ErrMissingAccountID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ledger.ValidateTransaction(tt.entries)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("ValidateTransaction() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}
		})
	}
}

func TestBuilders(t *testing.T) {
	t.Parallel()

	t.Run("NewBuyerDepositTransaction produces valid balanced entries", func(t *testing.T) {
		t.Parallel()

		entries, err := ledger.NewBuyerDepositTransaction(
			"tx_dep_01",
			"acc_escrow_1",
			"acc_buyer_dep_1",
			"trade_id_999",
			2500000, // $25,000.00
			"USD",
			"WIRE_REF_777888",
		)
		if err != nil {
			t.Fatalf("NewBuyerDepositTransaction failed: %v", err)
		}

		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		var sum int64
		for _, e := range entries {
			sum += e.AmountCents
		}
		if sum != 0 {
			t.Fatalf("expected zero sum, got %d", sum)
		}
	})

	t.Run("NewSettlementTransaction produces valid balanced 3-legged entries", func(t *testing.T) {
		t.Parallel()

		entries, err := ledger.NewSettlementTransaction(
			"tx_settle_01",
			"acc_escrow_1",
			"acc_seller_payout_1",
			"acc_verinode_rev_1",
			"trade_id_999",
			1000000, // $10,000 total
			50000,   // $500 fee
			"USD",
		)
		if err != nil {
			t.Fatalf("NewSettlementTransaction failed: %v", err)
		}

		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}

		var sum int64
		for _, e := range entries {
			sum += e.AmountCents
		}
		if sum != 0 {
			t.Fatalf("expected zero sum, got %d", sum)
		}
	})
}
