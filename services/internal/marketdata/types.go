package marketdata

import (
	"errors"
	"time"
)

var (
	ErrSeriesNotFound      = errors.New("index series not found")
	ErrInsufficientData    = errors.New("insufficient market data to calculate benchmark fix")
	ErrInvalidObservation  = errors.New("invalid market observation")
	ErrFlagNotFound        = errors.New("surveillance flag not found")
	ErrUnauthorizedSubject = errors.New("unauthorized surveillance subject")
)

type InputTier string

const (
	TierCompletedTrade      InputTier = "completed_trade"
	TierMatchedTradePending InputTier = "matched_trade_pending"
	TierFirmTwoSidedQuote   InputTier = "firm_two_sided_quote"
	TierFirmOneSidedQuote   InputTier = "firm_one_sided_quote"
	TierPublicListPrice     InputTier = "public_list_price"
)

// IndexSeries defines the physical specification for an index benchmark.
type IndexSeries struct {
	ID                   string    `json:"id"` // e.g. "H100-SXM-8XNV-US-WEEK-DEDICATED-USD"
	GPUModel             string    `json:"gpu_model"`
	Form                 string    `json:"form"`
	Topology             string    `json:"topology"`
	RegionBucket         string    `json:"region_bucket"`
	Tenor                string    `json:"tenor"`
	Tenancy              string    `json:"tenancy"`
	Currency             string    `json:"currency"`
	MethodologyVersion   string    `json:"methodology_version"`
	MinContributors      int       `json:"min_contributors"`
	MinNotionalUSD       float64   `json:"min_notional_usd"`
	MaxContributorWeight float64   `json:"max_contributor_weight"`
	CreatedAt            time.Time `json:"created_at"`
}

// MarketContribution represents a validated input into the index calculation window.
type MarketContribution struct {
	ID               string    `json:"id"`
	SeriesID         string    `json:"series_id"`
	Tier             InputTier `json:"tier"`
	ContractID       *string   `json:"contract_id,omitempty"`
	HourlyPriceUSD   float64   `json:"hourly_price_usd"`
	DurationHours    int       `json:"duration_hours"`
	NotionalUSD      float64   `json:"notional_usd"`
	ContributorID    string    `json:"contributor_id"`
	RelatedPartyFlag bool      `json:"related_party_flag"`
	CreatedAt        time.Time `json:"created_at"`
}

// IndexObservation represents a signed, published index fix.
type IndexObservation struct {
	ID                     string     `json:"id"`
	SeriesID               string     `json:"series_id"`
	Value                  *float64   `json:"value_usd,omitempty"` // nil if InsufficientData = true
	Unit                   string     `json:"unit"`
	ObservationWindowStart time.Time  `json:"observation_window_start"`
	ObservationWindowEnd   time.Time  `json:"observation_window_end"`
	PublishTime            time.Time  `json:"publish_time"`
	SequenceNumber         int64      `json:"sequence_number"`
	ContributorCount       int        `json:"contributor_count"`
	ObservationCount       int        `json:"observation_count"`
	TotalNotionalUSD       float64    `json:"total_notional_usd"`
	ConfidenceIntervalLow  *float64   `json:"confidence_interval_low,omitempty"`
	ConfidenceIntervalHigh *float64   `json:"confidence_interval_high,omitempty"`
	InsufficientData       bool       `json:"insufficient_data"`
	Reason                 *string    `json:"reason,omitempty"`
	Signature              *string    `json:"signature,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}

// SurveillanceFlag represents an anti-manipulation audit flag.
type SurveillanceFlag struct {
	ID          string                 `json:"id"`
	SubjectType string                 `json:"subject_type"` // "contract", "contribution", "participant"
	SubjectID   string                 `json:"subject_id"`
	FlagType    string                 `json:"flag_type"`    // "wash_trade", "related_party", "concentration", "spoofing", "end_window_marking"
	Severity    string                 `json:"severity"`     // "info", "warning", "critical"
	Details     map[string]interface{} `json:"details"`
	Status      string                 `json:"status"`       // "pending", "reviewed", "dismissed", "escalated"
	ReviewedBy  *string                `json:"reviewed_by,omitempty"`
	Resolution  *string                `json:"resolution,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}
