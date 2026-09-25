package contract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

type ContractRecord struct {
	ID              string    `json:"id"` // canonical trade_id
	RFQID           *string   `json:"rfq_id,omitempty"`
	QuoteID         *string   `json:"quote_id,omitempty"`
	BuyerID         string    `json:"buyer_id"`
	SellerID        string    `json:"seller_id"`
	GradeID         string    `json:"grade_id"`
	TemplateVersion string    `json:"template_version"`
	State           State     `json:"state"`
	SignedPDFHash   *string   `json:"signed_pdf_hash,omitempty"`
	SignedPDFRef    *string   `json:"signed_pdf_ref,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateContract(ctx context.Context, c *ContractRecord) error {
	if c.BuyerID == "" || c.SellerID == "" || c.GradeID == "" {
		return errors.New("buyer_id, seller_id, and grade_id are required")
	}
	if c.TemplateVersion == "" {
		c.TemplateVersion = "v1.0.0-institutional"
	}
	if c.State == "" {
		c.State = StateContractPending
	}

	query := `
		INSERT INTO contracts (rfq_id, quote_id, buyer_id, seller_id, grade_id, template_version, state, signed_pdf_hash, signed_pdf_ref)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at;
	`

	err := r.pool.QueryRow(ctx, query,
		c.RFQID,
		c.QuoteID,
		c.BuyerID,
		c.SellerID,
		c.GradeID,
		c.TemplateVersion,
		string(c.State),
		c.SignedPDFHash,
		c.SignedPDFRef,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("creating contract record: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, tradeID string) (*ContractRecord, error) {
	query := `
		SELECT id, rfq_id, quote_id, buyer_id, seller_id, grade_id, template_version, state, signed_pdf_hash, signed_pdf_ref, created_at, updated_at
		FROM contracts
		WHERE id = $1;
	`

	var c ContractRecord
	var st string
	err := r.pool.QueryRow(ctx, query, tradeID).Scan(
		&c.ID,
		&c.RFQID,
		&c.QuoteID,
		&c.BuyerID,
		&c.SellerID,
		&c.GradeID,
		&c.TemplateVersion,
		&st,
		&c.SignedPDFHash,
		&c.SignedPDFRef,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrContractNotFound
		}
		return nil, fmt.Errorf("getting contract %s: %w", tradeID, err)
	}

	state, err := ParseState(st)
	if err != nil {
		return nil, fmt.Errorf("parsing contract state %s: %w", st, err)
	}
	c.State = state

	return &c, nil
}

// AdvanceContractState enforces the Verinode 18-state transition machine in an atomic database transaction.
func (r *Repository) AdvanceContractState(ctx context.Context, tradeID string, nextState State, evt TransitionEvent) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transition transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Fetch current state with SELECT FOR UPDATE
	var currentStateStr string
	err = tx.QueryRow(ctx, "SELECT state FROM contracts WHERE id = $1 FOR UPDATE", tradeID).Scan(&currentStateStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrContractNotFound
		}
		return fmt.Errorf("locking contract: %w", err)
	}

	currentState, err := ParseState(currentStateStr)
	if err != nil {
		return fmt.Errorf("invalid existing state in DB: %w", err)
	}

	// 2. Validate transition through the state machine engine
	if err := ValidateTransition(currentState, nextState, evt); err != nil {
		return fmt.Errorf("state transition validation failed: %w", err)
	}

	authBytes, _ := json.Marshal(evt.AuthorizationDecision)

	// 3. Insert immutable contract audit event
	eventQuery := `
		INSERT INTO contract_events (contract_id, prior_state, new_state, actor, reason, idempotency_key, evidence_hash, authorization_decision)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err = tx.Exec(ctx, eventQuery,
		tradeID,
		string(currentState),
		string(nextState),
		evt.Actor,
		evt.Reason,
		evt.IdempotencyKey,
		evt.EvidenceHash,
		authBytes,
	)
	if err != nil {
		return fmt.Errorf("inserting contract event: %w", err)
	}

	// 4. Update contracts table state
	updateQuery := `
		UPDATE contracts 
		SET state = $1, updated_at = NOW() 
		WHERE id = $2;
	`
	if _, err := tx.Exec(ctx, updateQuery, string(nextState), tradeID); err != nil {
		return fmt.Errorf("updating contract state: %w", err)
	}

	// 5. Publish to event outbox
	outboxPayload, _ := json.Marshal(map[string]interface{}{
		"trade_id":    tradeID,
		"prior_state": currentState,
		"new_state":   nextState,
		"actor":       evt.Actor,
		"reason":      evt.Reason,
		"timestamp":   time.Now().UTC(),
	})
	outboxQuery := `
		INSERT INTO event_outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ('contract', $1, 'contract_state_changed', $2);
	`
	_, _ = tx.Exec(ctx, outboxQuery, tradeID, outboxPayload)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing state transition: %w", err)
	}

	return nil
}
