package participant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
	ErrInvalidParticipant  = errors.New("invalid participant data")
)

type Participant struct {
	ID               string    `json:"id"`
	LegalName        string    `json:"legal_name"`
	Jurisdiction     string    `json:"jurisdiction"`
	Role             string    `json:"role"` // 'buyer', 'seller', 'both'
	KYCStatus        string    `json:"kyc_status"` // 'pending', 'approved', 'rejected', 'review'
	SanctionsResult  []byte    `json:"sanctions_result,omitempty"`
	CreditLimitCents *int64    `json:"credit_limit_cents,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, p *Participant) error {
	if p.LegalName == "" || p.Jurisdiction == "" || p.Role == "" {
		return ErrInvalidParticipant
	}
	if p.KYCStatus == "" {
		p.KYCStatus = "pending"
	}

	query := `
		INSERT INTO participants (legal_name, jurisdiction, role, kyc_status, credit_limit_cents)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at;
	`

	err := r.pool.QueryRow(ctx, query,
		p.LegalName,
		p.Jurisdiction,
		p.Role,
		p.KYCStatus,
		p.CreditLimitCents,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("creating participant: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Participant, error) {
	query := `
		SELECT id, legal_name, jurisdiction, role, kyc_status, credit_limit_cents, created_at, updated_at
		FROM participants
		WHERE id = $1;
	`

	var p Participant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.LegalName,
		&p.Jurisdiction,
		&p.Role,
		&p.KYCStatus,
		&p.CreditLimitCents,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrParticipantNotFound
		}
		return nil, fmt.Errorf("getting participant %s: %w", id, err)
	}

	return &p, nil
}

func (r *Repository) List(ctx context.Context) ([]Participant, error) {
	query := `
		SELECT id, legal_name, jurisdiction, role, kyc_status, credit_limit_cents, created_at, updated_at
		FROM participants
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing participants: %w", err)
	}
	defer rows.Close()

	var result []Participant
	for rows.Next() {
		var p Participant
		if err := rows.Scan(
			&p.ID,
			&p.LegalName,
			&p.Jurisdiction,
			&p.Role,
			&p.KYCStatus,
			&p.CreditLimitCents,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning participant: %w", err)
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating participants: %w", err)
	}

	return result, nil
}
