package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAccountNotFound = errors.New("ledger account not found")
)

type AccountRecord struct {
	ID            string    `json:"id"`
	ParticipantID string    `json:"participant_id"`
	Currency      string    `json:"currency"`
	AccountType   string    `json:"account_type"` // 'deposit', 'payable', 'receivable', 'collateral'
	CreatedAt     time.Time `json:"created_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateAccount(ctx context.Context, a *AccountRecord) error {
	if a.ParticipantID == "" || a.AccountType == "" {
		return errors.New("participant_id and account_type are required")
	}
	if a.Currency == "" {
		a.Currency = "USD"
	}

	query := `
		INSERT INTO ledger_accounts (participant_id, currency, account_type)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query,
		a.ParticipantID,
		a.Currency,
		a.AccountType,
	).Scan(&a.ID, &a.CreatedAt)

	if err != nil {
		return fmt.Errorf("creating ledger account: %w", err)
	}

	return nil
}

// RecordTransaction verifies Invariant 3 (SUM(amount_cents) == 0) and records all entries atomically.
func (r *Repository) RecordTransaction(ctx context.Context, txID string, entries []Entry) error {
	if len(entries) < 2 {
		return ErrEmptyTransaction
	}

	// 1. Strictly validate double-entry balance in memory first
	if err := ValidateTransaction(entries); err != nil {
		return fmt.Errorf("double-entry validation rejected: %w", err)
	}

	// 2. Execute within an ACID transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning ledger transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	insertQuery := `
		INSERT INTO ledger_entries (transaction_id, account_id, contract_id, amount_cents, currency, bank_reference)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	for _, e := range entries {
		var contractID *string
		if e.ContractID != "" {
			contractID = &e.ContractID
		}
		var bankRef *string
		if e.BankReference != "" {
			bankRef = &e.BankReference
		}

		_, err := tx.Exec(ctx, insertQuery,
			txID,
			e.AccountID,
			contractID,
			e.AmountCents,
			e.Currency,
			bankRef,
		)
		if err != nil {
			return fmt.Errorf("inserting ledger entry for account %s: %w", e.AccountID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing ledger transaction: %w", err)
	}

	return nil
}

// GetAccountBalance returns the net sum of all entries for an account.
func (r *Repository) GetAccountBalance(ctx context.Context, accountID string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM ledger_entries
		WHERE account_id = $1;
	`

	var balance int64
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("calculating account balance %s: %w", accountID, err)
	}

	return balance, nil
}
