package claims

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrClaimNotFound = errors.New("claim not found")
	ErrInvalidClaim  = errors.New("invalid claim details")
)

type Claim struct {
	ID         string     `json:"id"`
	ContractID string     `json:"contract_id"`
	OpenedBy   string     `json:"opened_by"`
	Type       string     `json:"type"`  // 'outage', 'degradation', 'non_delivery', 'nonpayment'
	State      string     `json:"state"` // 'claim_open', 'resolved', 'disputed'
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type Evidence struct {
	ID           string    `json:"id"`
	ClaimID      string    `json:"claim_id"`
	EvidenceRef  string    `json:"evidence_ref"`
	EvidenceType string    `json:"evidence_type"`
	AddedBy      string    `json:"added_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, c *Claim) error {
	if c.ContractID == "" || c.OpenedBy == "" || c.Type == "" {
		return ErrInvalidClaim
	}
	if c.State == "" {
		c.State = "claim_open"
	}

	query := `
		INSERT INTO claims (contract_id, opened_by, type, state)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(ctx, query, c.ContractID, c.OpenedBy, c.Type, c.State).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating claim: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, stateFilter string) ([]Claim, error) {
	var query string
	var rows pgx.Rows
	var err error

	if stateFilter != "" {
		query = `
			SELECT id, contract_id, opened_by, type, state, created_at, resolved_at
			FROM claims
			WHERE state = $1
			ORDER BY created_at DESC;
		`
		rows, err = r.pool.Query(ctx, query, stateFilter)
	} else {
		query = `
			SELECT id, contract_id, opened_by, type, state, created_at, resolved_at
			FROM claims
			ORDER BY created_at DESC;
		`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("listing claims: %w", err)
	}
	defer rows.Close()

	var claims []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(
			&c.ID,
			&c.ContractID,
			&c.OpenedBy,
			&c.Type,
			&c.State,
			&c.CreatedAt,
			&c.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning claim: %w", err)
		}
		claims = append(claims, c)
	}
	return claims, nil
}

func (r *Repository) Resolve(ctx context.Context, claimID string, targetState string) error {
	query := `
		UPDATE claims
		SET state = $1, resolved_at = NOW()
		WHERE id = $2;
	`
	cmdTag, err := r.pool.Exec(ctx, query, targetState, claimID)
	if err != nil {
		return fmt.Errorf("resolving claim %s: %w", claimID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrClaimNotFound
	}
	return nil
}
