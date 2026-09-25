package rfq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRFQNotFound       = errors.New("rfq not found")
	ErrQuoteNotFound     = errors.New("quote not found")
	ErrInvalidRFQ        = errors.New("invalid rfq parameters")
	ErrInvalidQuote      = errors.New("invalid quote parameters")
	ErrQuoteExpired      = errors.New("quote has expired")
	ErrQuoteAlreadyTaken = errors.New("quote or rfq is no longer active")
	ErrUnauthorizedBuyer = errors.New("only the rfq buyer can accept this quote")
)

type RFQ struct {
	ID           string    `json:"id"`
	BuyerID      string    `json:"buyer_id"`
	GradeID      string    `json:"grade_id"`
	RegionBucket string    `json:"region_bucket"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Status       string    `json:"status"` // 'open', 'quoted', 'accepted', 'cancelled', 'expired'
	CreatedAt    time.Time `json:"created_at"`
}

type Quote struct {
	ID         string    `json:"id"`
	RFQID      string    `json:"rfq_id"`
	SellerID   string    `json:"seller_id"`
	BlockID    string    `json:"block_id"`
	PriceCents int64     `json:"price_cents"`
	Currency   string    `json:"currency"`
	ExpiresAt  time.Time `json:"expires_at"`
	Status     string    `json:"status"` // 'active', 'accepted', 'expired', 'withdrawn'
	CreatedAt  time.Time `json:"created_at"`
}

type AcceptedQuoteDetails struct {
	QuoteID    string
	RFQID      string
	BuyerID    string
	SellerID   string
	BlockID    string
	GradeID    string
	PriceCents int64
	Currency   string
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateRFQ(ctx context.Context, req *RFQ) error {
	if req.BuyerID == "" || req.GradeID == "" || req.RegionBucket == "" {
		return ErrInvalidRFQ
	}
	if !req.WindowEnd.After(req.WindowStart) {
		return fmt.Errorf("window_end must be after window_start")
	}
	if req.Status == "" {
		req.Status = "open"
	}

	query := `
		INSERT INTO rfqs (buyer_id, grade_id, region_bucket, window_start, window_end, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query,
		req.BuyerID,
		req.GradeID,
		req.RegionBucket,
		req.WindowStart,
		req.WindowEnd,
		req.Status,
	).Scan(&req.ID, &req.CreatedAt)

	if err != nil {
		return fmt.Errorf("creating rfq: %w", err)
	}

	return nil
}

func (r *Repository) GetRFQByID(ctx context.Context, id string) (*RFQ, error) {
	query := `
		SELECT id, buyer_id, grade_id, region_bucket, window_start, window_end, status, created_at
		FROM rfqs
		WHERE id = $1;
	`

	var req RFQ
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.BuyerID,
		&req.GradeID,
		&req.RegionBucket,
		&req.WindowStart,
		&req.WindowEnd,
		&req.Status,
		&req.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRFQNotFound
		}
		return nil, fmt.Errorf("getting rfq %s: %w", id, err)
	}

	return &req, nil
}

func (r *Repository) CreateQuote(ctx context.Context, q *Quote) error {
	if q.RFQID == "" || q.SellerID == "" || q.BlockID == "" || q.PriceCents <= 0 {
		return ErrInvalidQuote
	}
	if q.Currency == "" {
		q.Currency = "USD"
	}
	if q.Status == "" {
		q.Status = "active"
	}

	query := `
		INSERT INTO quotes (rfq_id, seller_id, block_id, price_cents, currency, expires_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query,
		q.RFQID,
		q.SellerID,
		q.BlockID,
		q.PriceCents,
		q.Currency,
		q.ExpiresAt,
		q.Status,
	).Scan(&q.ID, &q.CreatedAt)

	if err != nil {
		return fmt.Errorf("creating quote: %w", err)
	}

	// Update RFQ status to 'quoted' if currently 'open'
	_, _ = r.pool.Exec(ctx, "UPDATE rfqs SET status = 'quoted' WHERE id = $1 AND status = 'open'", q.RFQID)

	return nil
}

func (r *Repository) GetQuotesByRFQ(ctx context.Context, rfqID string) ([]Quote, error) {
	query := `
		SELECT id, rfq_id, seller_id, block_id, price_cents, currency, expires_at, status, created_at
		FROM quotes
		WHERE rfq_id = $1
		ORDER BY price_cents ASC;
	`

	rows, err := r.pool.Query(ctx, query, rfqID)
	if err != nil {
		return nil, fmt.Errorf("getting quotes for rfq %s: %w", rfqID, err)
	}
	defer rows.Close()

	var result []Quote
	for rows.Next() {
		var q Quote
		if err := rows.Scan(
			&q.ID,
			&q.RFQID,
			&q.SellerID,
			&q.BlockID,
			&q.PriceCents,
			&q.Currency,
			&q.ExpiresAt,
			&q.Status,
			&q.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning quote: %w", err)
		}
		result = append(result, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating quotes: %w", err)
	}

	return result, nil
}

// AcceptQuote atomically locks the Quote and RFQ, enforces concurrency safety (SELECT FOR UPDATE),
// transitions both to 'accepted', and marks inventory block 'reserved'.
func (r *Repository) AcceptQuote(ctx context.Context, quoteID string, buyerID string) (*AcceptedQuoteDetails, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning accept transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Lock and fetch Quote with SELECT FOR UPDATE
	quoteQuery := `
		SELECT q.id, q.rfq_id, q.seller_id, q.block_id, q.price_cents, q.currency, q.expires_at, q.status,
		       r.buyer_id, r.grade_id, r.status
		FROM quotes q
		JOIN rfqs r ON q.rfq_id = r.id
		WHERE q.id = $1
		FOR UPDATE OF q, r;
	`

	var (
		qID, rfqID, sellerID, blockID, currency string
		priceCents                              int64
		expiresAt                               time.Time
		quoteStatus, rfqBuyerID, gradeID, rfqStatus string
	)

	err = tx.QueryRow(ctx, quoteQuery, quoteID).Scan(
		&qID, &rfqID, &sellerID, &blockID, &priceCents, &currency, &expiresAt, &quoteStatus,
		&rfqBuyerID, &gradeID, &rfqStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQuoteNotFound
		}
		return nil, fmt.Errorf("locking quote: %w", err)
	}

	if rfqBuyerID != buyerID {
		return nil, ErrUnauthorizedBuyer
	}

	if time.Now().After(expiresAt) {
		return nil, ErrQuoteExpired
	}

	if quoteStatus != "active" || (rfqStatus != "open" && rfqStatus != "quoted") {
		return nil, ErrQuoteAlreadyTaken
	}

	// 2. Transition Quote to accepted
	if _, err := tx.Exec(ctx, "UPDATE quotes SET status = 'accepted' WHERE id = $1", qID); err != nil {
		return nil, fmt.Errorf("updating quote status: %w", err)
	}

	// 3. Transition RFQ to accepted
	if _, err := tx.Exec(ctx, "UPDATE rfqs SET status = 'accepted' WHERE id = $1", rfqID); err != nil {
		return nil, fmt.Errorf("updating rfq status: %w", err)
	}

	// 4. Reserve Inventory Block
	if _, err := tx.Exec(ctx, "UPDATE inventory_blocks SET status = 'reserved' WHERE id = $1", blockID); err != nil {
		return nil, fmt.Errorf("reserving inventory block: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing accept transaction: %w", err)
	}

	return &AcceptedQuoteDetails{
		QuoteID:    qID,
		RFQID:      rfqID,
		BuyerID:    rfqBuyerID,
		SellerID:   sellerID,
		BlockID:    blockID,
		GradeID:    gradeID,
		PriceCents: priceCents,
		Currency:   currency,
	}, nil
}
