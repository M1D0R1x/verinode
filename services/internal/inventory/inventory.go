package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBlockNotFound   = errors.New("inventory block not found")
	ErrInvalidBlock    = errors.New("invalid inventory block")
	ErrInvalidWindow   = errors.New("inventory block window_end must be after window_start")
)

type InventoryBlock struct {
	ID                   string     `json:"id"`
	SellerID             string     `json:"seller_id"`
	GradeID              string     `json:"grade_id"`
	RegionBucket         string     `json:"region_bucket"`
	FacilityRef          *string    `json:"facility_ref,omitempty"`
	WindowStart          time.Time  `json:"window_start"`
	WindowEnd            time.Time  `json:"window_end"`
	Status               string     `json:"status"` // 'available', 'reserved', 'delivered', 'released'
	LatestAttestationRef *string    `json:"latest_attestation_ref,omitempty"`
	LastHeartbeatAt      *time.Time `json:"last_heartbeat_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, b *InventoryBlock) error {
	if b.SellerID == "" || b.GradeID == "" || b.RegionBucket == "" {
		return ErrInvalidBlock
	}
	if !b.WindowEnd.After(b.WindowStart) {
		return ErrInvalidWindow
	}
	if b.Status == "" {
		b.Status = "available"
	}

	query := `
		INSERT INTO inventory_blocks (seller_id, grade_id, region_bucket, facility_ref, window_start, window_end, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at;
	`

	err := r.pool.QueryRow(ctx, query,
		b.SellerID,
		b.GradeID,
		b.RegionBucket,
		b.FacilityRef,
		b.WindowStart,
		b.WindowEnd,
		b.Status,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return fmt.Errorf("creating inventory block: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*InventoryBlock, error) {
	query := `
		SELECT id, seller_id, grade_id, region_bucket, facility_ref, window_start, window_end, status, latest_attestation_ref, last_heartbeat_at, created_at, updated_at
		FROM inventory_blocks
		WHERE id = $1;
	`

	var b InventoryBlock
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.SellerID,
		&b.GradeID,
		&b.RegionBucket,
		&b.FacilityRef,
		&b.WindowStart,
		&b.WindowEnd,
		&b.Status,
		&b.LatestAttestationRef,
		&b.LastHeartbeatAt,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBlockNotFound
		}
		return nil, fmt.Errorf("getting inventory block %s: %w", id, err)
	}

	return &b, nil
}

func (r *Repository) FindAvailableBlocks(ctx context.Context, gradeID, regionBucket string, start, end time.Time) ([]InventoryBlock, error) {
	query := `
		SELECT id, seller_id, grade_id, region_bucket, facility_ref, window_start, window_end, status, latest_attestation_ref, last_heartbeat_at, created_at, updated_at
		FROM inventory_blocks
		WHERE grade_id = $1
		  AND region_bucket = $2
		  AND status = 'available'
		  AND window_start <= $3
		  AND window_end >= $4
		ORDER BY window_start ASC;
	`

	rows, err := r.pool.Query(ctx, query, gradeID, regionBucket, start, end)
	if err != nil {
		return nil, fmt.Errorf("finding available blocks: %w", err)
	}
	defer rows.Close()

	var result []InventoryBlock
	for rows.Next() {
		var b InventoryBlock
		if err := rows.Scan(
			&b.ID,
			&b.SellerID,
			&b.GradeID,
			&b.RegionBucket,
			&b.FacilityRef,
			&b.WindowStart,
			&b.WindowEnd,
			&b.Status,
			&b.LatestAttestationRef,
			&b.LastHeartbeatAt,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning inventory block: %w", err)
		}
		result = append(result, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating inventory blocks: %w", err)
	}

	return result, nil
}
