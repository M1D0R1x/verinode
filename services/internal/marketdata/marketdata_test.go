package marketdata

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexCalculator_Invariant5(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	calc := NewCalculator(priv)

	series := &IndexSeries{
		ID:                   "H100-SXM-8XNV-US-WEEK-DEDICATED-USD",
		GPUModel:             "NVIDIA H100 SXM 80GB",
		Form:                 "8x SXM HGX",
		Topology:             "NVLink 4.0",
		RegionBucket:         "us-east",
		Tenor:                "168h",
		Tenancy:              "dedicated",
		Currency:             "USD",
		MethodologyVersion:   "v1.0.0-institutional",
		MinContributors:      3,
		MinNotionalUSD:       50000.0,
		MaxContributorWeight: 0.35,
	}

	windowStart := time.Now().Add(-24 * time.Hour)
	windowEnd := time.Now()

	t.Run("Invariant 5: Reject when contributor count is less than 3", func(t *testing.T) {
		contrib1 := "crusoe-energy-corp"
		contrib2 := "lambda-labs-inc"

		// Only 2 contributors, even with massive notional ($200,000)
		contributions := []MarketContribution{
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.50,
				DurationHours:  168,
				NotionalUSD:    100000.0,
				ContributorID:  contrib1,
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierMatchedTradePending,
				HourlyPriceUSD: 24.75,
				DurationHours:  168,
				NotionalUSD:    100000.0,
				ContributorID:  contrib2,
			},
		}

		obs := calc.CalculateFix(series, contributions, windowStart, windowEnd, 1)

		assert.True(t, obs.InsufficientData, "Invariant 5 MUST return insufficient_data: true when contributor threshold is unmet")
		assert.Nil(t, obs.Value, "Benchmark value must be nil when insufficient_data is true")
		assert.Equal(t, 2, obs.ContributorCount)
		assert.Contains(t, *obs.Reason, "Insufficient independent contributors")
	})

	t.Run("Invariant 5: Reject when total notional is below $50,000 floor", func(t *testing.T) {
		// 3 distinct contributors, but only $3,000 total notional
		contributions := []MarketContribution{
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.50,
				DurationHours:  24,
				NotionalUSD:    1000.0,
				ContributorID:  "seller-a",
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.60,
				DurationHours:  24,
				NotionalUSD:    1000.0,
				ContributorID:  "seller-b",
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.70,
				DurationHours:  24,
				NotionalUSD:    1000.0,
				ContributorID:  "seller-c",
			},
		}

		obs := calc.CalculateFix(series, contributions, windowStart, windowEnd, 2)

		assert.True(t, obs.InsufficientData, "Invariant 5 MUST return insufficient_data: true when notional is below liquidity floor")
		assert.Nil(t, obs.Value)
		assert.Equal(t, 3, obs.ContributorCount)
		assert.Contains(t, *obs.Reason, "below minimum liquidity floor")
	})

	t.Run("Compliant calculation with 3 independent contributors and >$50k volume", func(t *testing.T) {
		c1, c2, c3 := "crusoe-energy", "lambda-labs", "coreweave-inc"
		contributions := []MarketContribution{
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.00,
				DurationHours:  168,
				NotionalUSD:    25000.0,
				ContributorID:  c1,
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.50,
				DurationHours:  168,
				NotionalUSD:    35000.0,
				ContributorID:  c2,
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierFirmTwoSidedQuote,
				HourlyPriceUSD: 25.00,
				DurationHours:  168,
				NotionalUSD:    20000.0,
				ContributorID:  c3,
			},
		}

		obs := calc.CalculateFix(series, contributions, windowStart, windowEnd, 3)

		assert.False(t, obs.InsufficientData)
		require.NotNil(t, obs.Value)
		assert.Equal(t, 3, obs.ContributorCount)
		assert.GreaterOrEqual(t, *obs.Value, 24.00)
		assert.LessOrEqual(t, *obs.Value, 25.00)
		require.NotNil(t, obs.ConfidenceIntervalLow)
		require.NotNil(t, obs.ConfidenceIntervalHigh)
		assert.LessOrEqual(t, *obs.ConfidenceIntervalLow, *obs.Value)
		assert.GreaterOrEqual(t, *obs.ConfidenceIntervalHigh, *obs.Value)

		// Verify ED25519 signature
		require.NotNil(t, obs.Signature)
		sigBytes, err := hex.DecodeString(*obs.Signature)
		require.NoError(t, err)

		payload := fmt.Sprintf("%s:%d:%.4f:%s", obs.SeriesID, obs.SequenceNumber, *obs.Value, obs.PublishTime.Format(time.RFC3339))
		assert.True(t, ed25519.Verify(pub, []byte(payload), sigBytes), "Cryptographic index fix signature must be valid")
	})

	t.Run("Related-party contributions are strictly ignored", func(t *testing.T) {
		c1, c2 := "crusoe-energy", "lambda-labs"
		relatedC3 := "lambda-affiliate-fund"

		contributions := []MarketContribution{
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.00,
				DurationHours:  168,
				NotionalUSD:    50000.0,
				ContributorID:  c1,
			},
			{
				ID:             newUUID(),
				SeriesID:       series.ID,
				Tier:           TierCompletedTrade,
				HourlyPriceUSD: 24.50,
				DurationHours:  168,
				NotionalUSD:    50000.0,
				ContributorID:  c2,
			},
			{
				ID:               newUUID(),
				SeriesID:         series.ID,
				Tier:             TierCompletedTrade,
				HourlyPriceUSD:   30.00,
				DurationHours:    168,
				NotionalUSD:      50000.0,
				ContributorID:    relatedC3,
				RelatedPartyFlag: true, // Marked as related party
			},
		}

		obs := calc.CalculateFix(series, contributions, windowStart, windowEnd, 4)

		// After filtering related party, only 2 unique contributors remain -> insufficient data!
		assert.True(t, obs.InsufficientData)
		assert.Equal(t, 2, obs.ContributorCount)
	})
}
