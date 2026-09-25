package ledger

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnbalancedTransaction = errors.New("ledger: transaction entries do not balance to zero (SUM != 0)")
	ErrEmptyTransaction      = errors.New("ledger: transaction contains no entries")
	ErrMixedCurrencies       = errors.New("ledger: transaction entries have mixed currencies")
	ErrZeroAmountEntry       = errors.New("ledger: entry amount cannot be zero")
	ErrMissingAccountID      = errors.New("ledger: entry missing account_id")
	ErrMissingTransactionID  = errors.New("ledger: entry missing transaction_id")
)

// Entry represents an immutable atomic financial posting.
// By convention:
// - Positive amount_cents = DEBIT (Assets increase, Liabilities decrease)
// - Negative amount_cents = CREDIT (Assets decrease, Liabilities increase)
type Entry struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	AccountID     string    `json:"account_id"`
	ContractID    string    `json:"contract_id,omitempty"` // canonical trade_id
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
	BankReference string    `json:"bank_reference,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ValidateTransaction enforces the double-entry invariant:
// SUM(amount_cents) == 0 grouped by transaction_id and currency.
func ValidateTransaction(entries []Entry) error {
	if len(entries) < 2 {
		return fmt.Errorf("%w: a double-entry transaction must have at least 2 entries, got %d", ErrEmptyTransaction, len(entries))
	}

	txID := entries[0].TransactionID
	if txID == "" {
		return ErrMissingTransactionID
	}

	currency := entries[0].Currency
	if currency == "" {
		return errors.New("ledger: missing currency")
	}

	var sum int64 = 0

	for i, e := range entries {
		if e.TransactionID != txID {
			return fmt.Errorf("ledger: entry %d has mismatched transaction_id %q (expected %q)", i, e.TransactionID, txID)
		}
		if e.AccountID == "" {
			return fmt.Errorf("%w at index %d", ErrMissingAccountID, i)
		}
		if e.Currency != currency {
			return fmt.Errorf("%w: entry %d has currency %q (expected %q)", ErrMixedCurrencies, i, e.Currency, currency)
		}
		if e.AmountCents == 0 {
			return fmt.Errorf("%w at index %d", ErrZeroAmountEntry, i)
		}

		sum += e.AmountCents
	}

	if sum != 0 {
		return fmt.Errorf("%w: net balance is %d cents (must be 0)", ErrUnbalancedTransaction, sum)
	}

	return nil
}
