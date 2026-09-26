package marketdata

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// Calculator computes robust volume-weighted index fixes with concentration guards.
type Calculator struct {
	signingKey ed25519.PrivateKey
}

func NewCalculator(signingKey ed25519.PrivateKey) *Calculator {
	return &Calculator{signingKey: signingKey}
}

type weightedPoint struct {
	price         float64
	weight        float64
	contributorID string
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// CalculateFix executes the canonical benchmark methodology strictly adhering to Invariant 5.
func (c *Calculator) CalculateFix(
	series *IndexSeries,
	contributions []MarketContribution,
	windowStart, windowEnd time.Time,
	seqNum int64,
) IndexObservation {
	now := time.Now().UTC()
	obs := IndexObservation{
		ID:                     newUUID(),
		SeriesID:               series.ID,
		Unit:                   "USD_PER_NODE_HOUR",
		ObservationWindowStart: windowStart,
		ObservationWindowEnd:   windowEnd,
		PublishTime:            now,
		SequenceNumber:         seqNum,
		ObservationCount:       len(contributions),
		CreatedAt:              now,
	}

	if len(contributions) == 0 {
		obs.InsufficientData = true
		reason := "No market contributions recorded in observation window"
		obs.Reason = &reason
		return obs
	}

	uniqueContributors := make(map[string]bool)
	var totalNotional float64
	for _, contrib := range contributions {
		// Filter out related-party contributions or invalid duration
		if contrib.RelatedPartyFlag || contrib.DurationHours <= 0 {
			continue
		}
		uniqueContributors[contrib.ContributorID] = true
		totalNotional += contrib.NotionalUSD
	}

	obs.ContributorCount = len(uniqueContributors)
	obs.TotalNotionalUSD = totalNotional

	// Invariant 5 Enforcement 1: Minimum Contributor Count Threshold
	if obs.ContributorCount < series.MinContributors {
		obs.InsufficientData = true
		reason := fmt.Sprintf("Insufficient independent contributors: found %d, minimum required is %d",
			obs.ContributorCount, series.MinContributors)
		obs.Reason = &reason
		return obs
	}

	// Invariant 5 Enforcement 2: Minimum Notional Volume Threshold
	if totalNotional < series.MinNotionalUSD {
		obs.InsufficientData = true
		reason := fmt.Sprintf("Insufficient volume: $%.2f notional is below minimum liquidity floor $%.2f",
			totalNotional, series.MinNotionalUSD)
		obs.Reason = &reason
		return obs
	}

	// Weight assignment by input hierarchy tier
	tierWeights := map[InputTier]float64{
		TierCompletedTrade:      1.0,
		TierMatchedTradePending: 0.8,
		TierFirmTwoSidedQuote:   0.5,
		TierFirmOneSidedQuote:   0.25,
		TierPublicListPrice:     0.1,
	}

	var points []weightedPoint
	contribTotals := make(map[string]float64)

	for _, contrib := range contributions {
		if contrib.RelatedPartyFlag || contrib.DurationHours <= 0 {
			continue
		}
		tw, ok := tierWeights[contrib.Tier]
		if !ok {
			tw = 0.2
		}
		rawWeight := contrib.NotionalUSD * tw
		points = append(points, weightedPoint{
			price:         contrib.HourlyPriceUSD,
			weight:        rawWeight,
			contributorID: contrib.ContributorID,
		})
		contribTotals[contrib.ContributorID] += rawWeight
	}

	if len(points) == 0 {
		obs.InsufficientData = true
		reason := "Zero eligible contributions after excluding related parties"
		obs.Reason = &reason
		return obs
	}

	// Concentration Guard: Cap any contributor's weight to MaxContributorWeight (e.g. 35%)
	maxWeightFraction := series.MaxContributorWeight
	if maxWeightFraction <= 0 || maxWeightFraction >= 1.0 {
		maxWeightFraction = 0.35
	}
	ratioCap := maxWeightFraction / (1.0 - maxWeightFraction)

	scaling := make(map[string]float64)
	for contribID, totalW := range contribTotals {
		var otherWeights float64
		for otherID, otherW := range contribTotals {
			if otherID != contribID {
				otherWeights += otherW
			}
		}
		maxAllowed := otherWeights * ratioCap
		if totalW > maxAllowed && totalW > 0 {
			scaling[contribID] = maxAllowed / totalW
		} else {
			scaling[contribID] = 1.0
		}
	}

	// Apply scaling to points
	var sumWeights float64
	for i := range points {
		factor := scaling[points[i].contributorID]
		points[i].weight *= factor
		sumWeights += points[i].weight
	}

	if sumWeights <= 0 {
		obs.InsufficientData = true
		reason := "Weights collapsed after concentration scaling"
		obs.Reason = &reason
		return obs
	}

	// Sort points ascending by price for cumulative volume-weighted median & percentiles
	sort.Slice(points, func(i, j int) bool {
		return points[i].price < points[j].price
	})

	medianThreshold := sumWeights * 0.50
	p25Threshold := sumWeights * 0.25
	p75Threshold := sumWeights * 0.75

	var cumWeight float64
	var medianPrice, p25Price, p75Price float64
	p25Found, medianFound, p75Found := false, false, false

	for _, pt := range points {
		cumWeight += pt.weight
		if !p25Found && cumWeight >= p25Threshold {
			p25Price = pt.price
			p25Found = true
		}
		if !medianFound && cumWeight >= medianThreshold {
			medianPrice = pt.price
			medianFound = true
		}
		if !p75Found && cumWeight >= p75Threshold {
			p75Price = pt.price
			p75Found = true
		}
	}

	// Default fallback to highest point if floating rounding edge
	if !p25Found && len(points) > 0 {
		p25Price = points[0].price
	}
	if !medianFound && len(points) > 0 {
		medianPrice = points[len(points)/2].price
	}
	if !p75Found && len(points) > 0 {
		p75Price = points[len(points)-1].price
	}

	obs.Value = &medianPrice
	obs.ConfidenceIntervalLow = &p25Price
	obs.ConfidenceIntervalHigh = &p75Price
	obs.InsufficientData = false

	// Sign the canonical fix envelope if signing key is present
	if len(c.signingKey) == ed25519.PrivateKeySize {
		payload := fmt.Sprintf("%s:%d:%.4f:%s", obs.SeriesID, obs.SequenceNumber, medianPrice, obs.PublishTime.Format(time.RFC3339))
		sig := ed25519.Sign(c.signingKey, []byte(payload))
		sigHex := hex.EncodeToString(sig)
		obs.Signature = &sigHex
	} else {
		// Mock signature using sha256
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%.4f", obs.SeriesID, obs.SequenceNumber, medianPrice)))
		hashHex := hex.EncodeToString(hash[:])
		obs.Signature = &hashHex
	}

	return obs
}
