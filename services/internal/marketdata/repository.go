package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetSeries retrieves series by its identifier.
func (r *Repository) GetSeries(ctx context.Context, id string) (*IndexSeries, error) {
	query := `
		SELECT id, gpu_model, form, topology, region_bucket, tenor, tenancy, currency,
		       methodology_version, min_contributors, min_notional_usd, max_contributor_weight, created_at
		FROM index_series
		WHERE id = $1;
	`
	var s IndexSeries
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.GPUModel, &s.Form, &s.Topology, &s.RegionBucket, &s.Tenor, &s.Tenancy, &s.Currency,
		&s.MethodologyVersion, &s.MinContributors, &s.MinNotionalUSD, &s.MaxContributorWeight, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSeriesNotFound
		}
		return nil, fmt.Errorf("query series failed: %w", err)
	}
	return &s, nil
}

// ListSeries lists all configured index series.
func (r *Repository) ListSeries(ctx context.Context) ([]IndexSeries, error) {
	query := `
		SELECT id, gpu_model, form, topology, region_bucket, tenor, tenancy, currency,
		       methodology_version, min_contributors, min_notional_usd, max_contributor_weight, created_at
		FROM index_series
		ORDER BY id ASC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list series failed: %w", err)
	}
	defer rows.Close()

	var result []IndexSeries
	for rows.Next() {
		var s IndexSeries
		if err := rows.Scan(
			&s.ID, &s.GPUModel, &s.Form, &s.Topology, &s.RegionBucket, &s.Tenor, &s.Tenancy, &s.Currency,
			&s.MethodologyVersion, &s.MinContributors, &s.MinNotionalUSD, &s.MaxContributorWeight, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// UpsertSeries creates or updates an index series definition.
func (r *Repository) UpsertSeries(ctx context.Context, s *IndexSeries) error {
	query := `
		INSERT INTO index_series (
			id, gpu_model, form, topology, region_bucket, tenor, tenancy, currency,
			methodology_version, min_contributors, min_notional_usd, max_contributor_weight, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (id) DO UPDATE SET
			methodology_version = EXCLUDED.methodology_version,
			min_contributors = EXCLUDED.min_contributors,
			min_notional_usd = EXCLUDED.min_notional_usd,
			max_contributor_weight = EXCLUDED.max_contributor_weight;
	`
	_, err := r.pool.Exec(ctx, query,
		s.ID, s.GPUModel, s.Form, s.Topology, s.RegionBucket, s.Tenor, s.Tenancy, s.Currency,
		s.MethodologyVersion, s.MinContributors, s.MinNotionalUSD, s.MaxContributorWeight,
	)
	return err
}

// GetLatestObservation gets the most recent published fix for a series.
func (r *Repository) GetLatestObservation(ctx context.Context, seriesID string) (*IndexObservation, error) {
	query := `
		SELECT id, series_id, value, unit, observation_window_start, observation_window_end,
		       publish_time, sequence_number, contributor_count, observation_count, total_notional_usd,
		       confidence_interval_low, confidence_interval_high, insufficient_data, reason, signature, created_at
		FROM index_observations
		WHERE series_id = $1
		ORDER BY publish_time DESC, sequence_number DESC
		LIMIT 1;
	`
	var obs IndexObservation
	err := r.pool.QueryRow(ctx, query, seriesID).Scan(
		&obs.ID, &obs.SeriesID, &obs.Value, &obs.Unit, &obs.ObservationWindowStart, &obs.ObservationWindowEnd,
		&obs.PublishTime, &obs.SequenceNumber, &obs.ContributorCount, &obs.ObservationCount, &obs.TotalNotionalUSD,
		&obs.ConfidenceIntervalLow, &obs.ConfidenceIntervalHigh, &obs.InsufficientData, &obs.Reason, &obs.Signature, &obs.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No observation yet
		}
		return nil, fmt.Errorf("get latest observation failed: %w", err)
	}
	return &obs, nil
}

// ListObservations returns historical observations for a series.
func (r *Repository) ListObservations(ctx context.Context, seriesID string, limit int) ([]IndexObservation, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT id, series_id, value, unit, observation_window_start, observation_window_end,
		       publish_time, sequence_number, contributor_count, observation_count, total_notional_usd,
		       confidence_interval_low, confidence_interval_high, insufficient_data, reason, signature, created_at
		FROM index_observations
		WHERE series_id = $1
		ORDER BY publish_time DESC, sequence_number DESC
		LIMIT $2;
	`
	rows, err := r.pool.Query(ctx, query, seriesID, limit)
	if err != nil {
		return nil, fmt.Errorf("list observations failed: %w", err)
	}
	defer rows.Close()

	var result []IndexObservation
	for rows.Next() {
		var obs IndexObservation
		if err := rows.Scan(
			&obs.ID, &obs.SeriesID, &obs.Value, &obs.Unit, &obs.ObservationWindowStart, &obs.ObservationWindowEnd,
			&obs.PublishTime, &obs.SequenceNumber, &obs.ContributorCount, &obs.ObservationCount, &obs.TotalNotionalUSD,
			&obs.ConfidenceIntervalLow, &obs.ConfidenceIntervalHigh, &obs.InsufficientData, &obs.Reason, &obs.Signature, &obs.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, obs)
	}
	return result, rows.Err()
}

// RecordObservation saves a newly calculated benchmark fix.
func (r *Repository) RecordObservation(ctx context.Context, obs *IndexObservation) error {
	if obs.ID == "" {
		obs.ID = newUUID()
	}
	if obs.CreatedAt.IsZero() {
		obs.CreatedAt = time.Now().UTC()
	}
	query := `
		INSERT INTO index_observations (
			id, series_id, value, unit, observation_window_start, observation_window_end,
			publish_time, sequence_number, contributor_count, observation_count, total_notional_usd,
			confidence_interval_low, confidence_interval_high, insufficient_data, reason, signature, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (series_id, sequence_number) DO UPDATE SET
			value = EXCLUDED.value,
			insufficient_data = EXCLUDED.insufficient_data,
			signature = EXCLUDED.signature;
	`
	_, err := r.pool.Exec(ctx, query,
		obs.ID, obs.SeriesID, obs.Value, obs.Unit, obs.ObservationWindowStart, obs.ObservationWindowEnd,
		obs.PublishTime, obs.SequenceNumber, obs.ContributorCount, obs.ObservationCount, obs.TotalNotionalUSD,
		obs.ConfidenceIntervalLow, obs.ConfidenceIntervalHigh, obs.InsufficientData, obs.Reason, obs.Signature, obs.CreatedAt,
	)
	return err
}

// RecordContribution registers an input contribution from trade or quotes.
func (r *Repository) RecordContribution(ctx context.Context, c *MarketContribution) error {
	if c.ID == "" {
		c.ID = newUUID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	query := `
		INSERT INTO index_contributions (
			id, series_id, tier, contract_id, hourly_price_usd, duration_hours,
			notional_usd, contributor_id, related_party_flag, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	_, err := r.pool.Exec(ctx, query,
		c.ID, c.SeriesID, string(c.Tier), c.ContractID, c.HourlyPriceUSD, c.DurationHours,
		c.NotionalUSD, c.ContributorID, c.RelatedPartyFlag, c.CreatedAt,
	)
	return err
}

// ListContributions fetches contributions for an observation window.
func (r *Repository) ListContributions(ctx context.Context, seriesID string, since time.Time) ([]MarketContribution, error) {
	query := `
		SELECT id, series_id, tier, contract_id, hourly_price_usd, duration_hours,
		       notional_usd, contributor_id, related_party_flag, created_at
		FROM index_contributions
		WHERE series_id = $1 AND created_at >= $2
		ORDER BY created_at ASC;
	`
	rows, err := r.pool.Query(ctx, query, seriesID, since)
	if err != nil {
		return nil, fmt.Errorf("list contributions failed: %w", err)
	}
	defer rows.Close()

	var result []MarketContribution
	for rows.Next() {
		var c MarketContribution
		var tierStr string
		if err := rows.Scan(
			&c.ID, &c.SeriesID, &tierStr, &c.ContractID, &c.HourlyPriceUSD, &c.DurationHours,
			&c.NotionalUSD, &c.ContributorID, &c.RelatedPartyFlag, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		c.Tier = InputTier(tierStr)
		result = append(result, c)
	}
	return result, rows.Err()
}

// ListSurveillanceFlags lists anti-manipulation flags.
func (r *Repository) ListSurveillanceFlags(ctx context.Context, status string) ([]SurveillanceFlag, error) {
	query := `
		SELECT id, subject_type, subject_id, flag_type, severity, details, status, reviewed_by, resolution, created_at, resolved_at
		FROM surveillance_flags
	`
	var rows pgx.Rows
	var err error
	if status != "" && status != "all" {
		query += ` WHERE status = $1 ORDER BY created_at DESC;`
		rows, err = r.pool.Query(ctx, query, status)
	} else {
		query += ` ORDER BY created_at DESC;`
		rows, err = r.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("list surveillance flags failed: %w", err)
	}
	defer rows.Close()

	var result []SurveillanceFlag
	for rows.Next() {
		var f SurveillanceFlag
		var detailsJSON []byte
		if err := rows.Scan(
			&f.ID, &f.SubjectType, &f.SubjectID, &f.FlagType, &f.Severity, &detailsJSON, &f.Status,
			&f.ReviewedBy, &f.Resolution, &f.CreatedAt, &f.ResolvedAt,
		); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &f.Details)
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// CreateSurveillanceFlag inserts a new flag.
func (r *Repository) CreateSurveillanceFlag(ctx context.Context, f *SurveillanceFlag) error {
	if f.ID == "" {
		f.ID = newUUID()
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	detailsJSON, _ := json.Marshal(f.Details)

	query := `
		INSERT INTO surveillance_flags (
			id, subject_type, subject_id, flag_type, severity, details, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err := r.pool.Exec(ctx, query,
		f.ID, f.SubjectType, f.SubjectID, f.FlagType, f.Severity, detailsJSON, f.Status, f.CreatedAt,
	)
	return err
}

// ReviewSurveillanceFlag records an audit decision on a surveillance flag.
func (r *Repository) ReviewSurveillanceFlag(ctx context.Context, id, reviewer, resolution, newStatus string) error {
	now := time.Now().UTC()
	query := `
		UPDATE surveillance_flags
		SET status = $1, reviewed_by = $2, resolution = $3, resolved_at = $4
		WHERE id = $5;
	`
	cmdTag, err := r.pool.Exec(ctx, query, newStatus, reviewer, resolution, now, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrFlagNotFound
	}
	return nil
}
