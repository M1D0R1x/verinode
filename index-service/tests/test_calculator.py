"""
Unit Tests for Verinode Index Calculation Engine & Invariant 5 Enforcement
Compatible with both unittest and pytest runners.
"""

import sys
import os
import unittest
from datetime import datetime, timezone

# Add src to sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "../src")))

from verinode_index.models import MarketObservation, InputTier
from verinode_index.calculator import IndexCalculator
from verinode_index.publisher import IndexPublisher


class TestIndexCalculator(unittest.TestCase):
    def setUp(self):
        self.calculator = IndexCalculator(
            min_contributors=3,
            min_notional_usd=50_000.0,
            max_contributor_weight=0.35,
        )

    def test_successful_benchmark_calculation(self):
        now = datetime.now(timezone.utc)
        series_id = "H100-SXM-8XNV-168H-US-DEDICATED-USD"

        observations = [
            MarketObservation(
                contributor_id="contrib_apex",
                hourly_price_usd=210.0,
                duration_hours=168,
                tier=InputTier.COMPLETED_TRADE,
                timestamp=now,
            ),
            MarketObservation(
                contributor_id="contrib_nebula",
                hourly_price_usd=215.0,
                duration_hours=168,
                tier=InputTier.COMPLETED_TRADE,
                timestamp=now,
            ),
            MarketObservation(
                contributor_id="contrib_vantage",
                hourly_price_usd=218.0,
                duration_hours=168,
                tier=InputTier.MATCHED_TRADE,
                timestamp=now,
            ),
        ]

        result = self.calculator.calculate_benchmark(series_id, observations)

        self.assertFalse(result.insufficient_data)
        self.assertIsNotNone(result.value_usd)
        self.assertTrue(210.0 <= result.value_usd <= 218.0)
        self.assertEqual(result.contributor_count, 3)
        self.assertEqual(result.observation_count, 3)
        self.assertGreater(result.total_notional_usd, 50_000.0)

        artifact = IndexPublisher.format_publication_artifact(result)
        self.assertEqual(artifact["series"], series_id)
        self.assertFalse(artifact["insufficient_data"])
        self.assertEqual(artifact["benchmark_price_usd"], result.value_usd)

    def test_invariant_5_too_few_contributors(self):
        """Invariant 5: Return insufficient_data: true when minimum contributor counts are unmet."""
        now = datetime.now(timezone.utc)
        series_id = "H100-SXM-8XNV-168H-US-DEDICATED-USD"

        observations = [
            MarketObservation(
                contributor_id="contrib_apex",
                hourly_price_usd=210.0,
                duration_hours=336,
                tier=InputTier.COMPLETED_TRADE,
                timestamp=now,
            ),
            MarketObservation(
                contributor_id="contrib_nebula",
                hourly_price_usd=212.0,
                duration_hours=336,
                tier=InputTier.COMPLETED_TRADE,
                timestamp=now,
            ),
        ]

        result = self.calculator.calculate_benchmark(series_id, observations)

        self.assertTrue(result.insufficient_data)
        self.assertIsNone(result.value_usd)
        self.assertIn("Insufficient independent contributors", result.reason or "")
        self.assertEqual(result.contributor_count, 2)

    def test_invariant_5_below_minimum_volume(self):
        """Invariant 5: Return insufficient_data: true when volume thresholds are unmet."""
        now = datetime.now(timezone.utc)
        series_id = "H100-SXM-8XNV-168H-US-DEDICATED-USD"

        observations = [
            MarketObservation("contrib_a", 210.0, 1, InputTier.COMPLETED_TRADE, now),
            MarketObservation("contrib_b", 210.0, 1, InputTier.COMPLETED_TRADE, now),
            MarketObservation("contrib_c", 210.0, 1, InputTier.COMPLETED_TRADE, now),
        ]

        result = self.calculator.calculate_benchmark(series_id, observations)

        self.assertTrue(result.insufficient_data)
        self.assertIsNone(result.value_usd)
        self.assertIn("Insufficient volume", result.reason or "")

    def test_invariant_5_empty_dataset(self):
        """Invariant 5: Empty dataset must return insufficient_data: true, never default or fabricated value."""
        result = self.calculator.calculate_benchmark("H100-SXM-8XNV-168H-US-DEDICATED-USD", [])
        self.assertTrue(result.insufficient_data)
        self.assertIsNone(result.value_usd)
        self.assertEqual(result.contributor_count, 0)

    def test_contributor_concentration_clamping(self):
        """Ensures a single dominant supplier cannot dictate the index median."""
        now = datetime.now(timezone.utc)
        series_id = "H100-SXM-8XNV-168H-US-DEDICATED-USD"

        observations = [
            MarketObservation("dominant_whale", 300.0, 1000, InputTier.COMPLETED_TRADE, now),
            MarketObservation("contrib_b", 210.0, 168, InputTier.COMPLETED_TRADE, now),
            MarketObservation("contrib_c", 212.0, 168, InputTier.COMPLETED_TRADE, now),
        ]

        result = self.calculator.calculate_benchmark(series_id, observations)

        self.assertFalse(result.insufficient_data)
        self.assertIsNotNone(result.value_usd)
        self.assertLess(result.value_usd, 300.0)


if __name__ == "__main__":
    unittest.main()
