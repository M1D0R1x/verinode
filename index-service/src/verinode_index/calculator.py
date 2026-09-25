"""
Verinode Institutional Index Calculation Engine
Calculates robust volume-weighted benchmark fixes with concentration guards.
Strictly adheres to Invariant 5: Return insufficient_data: true when thresholds are unmet.
"""

from datetime import datetime, timezone
from typing import List, Tuple
from .models import MarketObservation, IndexFixResult, InputTier


class IndexCalculator:
    def __init__(
        self,
        min_contributors: int = 3,
        min_notional_usd: float = 50_000.0,
        max_contributor_weight: float = 0.35, # 35% concentration cap
    ):
        self.min_contributors = min_contributors
        self.min_notional_usd = min_notional_usd
        self.max_contributor_weight = max_contributor_weight

    def calculate_benchmark(
        self,
        series_id: str,
        observations: List[MarketObservation],
    ) -> IndexFixResult:
        now_str = datetime.now(timezone.utc).isoformat()

        if not observations:
            return IndexFixResult(
                series_id=series_id,
                unit="USD_PER_NODE_HOUR",
                insufficient_data=True,
                contributor_count=0,
                observation_count=0,
                total_notional_usd=0.0,
                publish_time=now_str,
                reason="No market observations available for observation window",
            )

        unique_contributors = {obs.contributor_id for obs in observations}
        total_notional = sum(obs.notional_usd for obs in observations)

        # Invariant 5 Enforcement 1: Minimum Contributor Threshold
        if len(unique_contributors) < self.min_contributors:
            return IndexFixResult(
                series_id=series_id,
                unit="USD_PER_NODE_HOUR",
                insufficient_data=True,
                contributor_count=len(unique_contributors),
                observation_count=len(observations),
                total_notional_usd=total_notional,
                publish_time=now_str,
                reason=f"Insufficient independent contributors: found {len(unique_contributors)}, required at least {self.min_contributors}",
            )

        # Invariant 5 Enforcement 2: Minimum Notional Volume Threshold
        if total_notional < self.min_notional_usd:
            return IndexFixResult(
                series_id=series_id,
                unit="USD_PER_NODE_HOUR",
                insufficient_data=True,
                contributor_count=len(unique_contributors),
                observation_count=len(observations),
                total_notional_usd=total_notional,
                publish_time=now_str,
                reason=f"Insufficient volume: ${total_notional:,.2f} below minimum liquidity threshold ${self.min_notional_usd:,.2f}",
            )

        # Calculate Tier & Notional Weighted Values
        weighted_points: List[Tuple[float, float, str]] = [] # (price, raw_weight, contributor_id)

        tier_multipliers = {
            InputTier.COMPLETED_TRADE: 1.0,
            InputTier.MATCHED_TRADE: 0.8,
            InputTier.FIRM_TWO_SIDED_QUOTE: 0.5,
            InputTier.FIRM_ONE_SIDED_QUOTE: 0.25,
        }

        for obs in observations:
            tier_weight = tier_multipliers.get(obs.tier, 0.2)
            raw_weight = obs.notional_usd * tier_weight
            weighted_points.append((obs.hourly_price_usd, raw_weight, obs.contributor_id))

        # Enforce Contributor Concentration Cap (max 35% of final weight)
        # Iteratively cap any contributor whose weight exceeds c / (1 - c) * sum(others)
        contributor_totals = {}
        for _, weight, contrib in weighted_points:
            contributor_totals[contrib] = contributor_totals.get(contrib, 0.0) + weight

        c = self.max_contributor_weight
        ratio_cap = c / (1.0 - c) if c < 1.0 else float("inf")

        scaling_factors = {contrib: 1.0 for contrib in contributor_totals}
        for contrib, total_w in contributor_totals.items():
            other_weights = sum(w for other, w in contributor_totals.items() if other != contrib)
            max_allowed = other_weights * ratio_cap
            if total_w > max_allowed:
                scaling_factors[contrib] = max_allowed / total_w

        clamped_points: List[Tuple[float, float]] = []
        for price, weight, contrib in weighted_points:
            adjusted_weight = weight * scaling_factors[contrib]
            clamped_points.append((price, adjusted_weight))

        # Compute Volume-Weighted Median
        clamped_points.sort(key=lambda x: x[0])
        total_adj_weight = sum(w for _, w in clamped_points)
        cumulative_target = total_adj_weight / 2.0

        current_cumulative = 0.0
        median_price = clamped_points[0][0]

        for price, weight in clamped_points:
            current_cumulative += weight
            if current_cumulative >= cumulative_target:
                median_price = price
                break

        # Compute Dispersion (P25 and P75 bounds)
        p25_target = total_adj_weight * 0.25
        p75_target = total_adj_weight * 0.75
        p25_price = clamped_points[0][0]
        p75_price = clamped_points[-1][0]

        curr = 0.0
        for price, weight in clamped_points:
            curr += weight
            if curr >= p25_target and p25_price == clamped_points[0][0]:
                p25_price = price
            if curr >= p75_target:
                p75_price = price
                break

        return IndexFixResult(
            series_id=series_id,
            unit="USD_PER_NODE_HOUR",
            insufficient_data=False,
            value_usd=round(median_price, 2),
            confidence_interval_low=round(p25_price, 2),
            confidence_interval_high=round(p75_price, 2),
            contributor_count=len(unique_contributors),
            observation_count=len(observations),
            total_notional_usd=round(total_notional, 2),
            publish_time=now_str,
            reason="Volume-weighted median fix computed successfully with concentration limits applied",
        )
