import unittest
from datetime import datetime, timezone
from verinode_index.models import IndexFixResult
from verinode_index.hyperliquid_publisher import HyperliquidOraclePublisher


class TestHyperliquidOraclePublisher(unittest.TestCase):
    def setUp(self):
        self.publisher = HyperliquidOraclePublisher()

    def test_format_hip3_oracle_envelope_success(self):
        fix = IndexFixResult(
            series_id="VN-H100-SXM-168H",
            unit="USD_PER_GPU_HOUR",
            publish_time=datetime.now(timezone.utc).isoformat(),
            value_usd=2.25,
            confidence_interval_low=2.15,
            confidence_interval_high=2.35,
            contributor_count=4,
            observation_count=12,
            total_notional_usd=250000.0,
            insufficient_data=False,
            methodology_version="v1.0.0-vw-median",
            reason=None,
        )

        envelope = self.publisher.format_hip3_oracle_envelope(fix)
        self.assertEqual(envelope["type"], "hip3_oracle_update")
        self.assertEqual(envelope["mark_price"], 2.25)
        self.assertEqual(envelope["contributor_count"], 4)
        self.assertEqual(envelope["status"], "ACTIVE_FEED")

    def test_invariant_5_rejects_insufficient_data_publication(self):
        fix = IndexFixResult(
            series_id="VN-H100-SXM-168H",
            unit="USD_PER_GPU_HOUR",
            publish_time=datetime.now(timezone.utc).isoformat(),
            value_usd=None,
            confidence_interval_low=None,
            confidence_interval_high=None,
            contributor_count=1,
            observation_count=2,
            total_notional_usd=15000.0,
            insufficient_data=True,
            methodology_version="v1.0.0-vw-median",
            reason="Minimum 3 distinct contributors required (got 1)",
        )

        with self.assertRaises(ValueError) as ctx:
            self.publisher.format_hip3_oracle_envelope(fix)

        self.assertIn("Invariant 5 Violation Prevented", str(ctx.exception))

    def test_compute_forward_hedge_quote(self):
        # 168 hours * 8 GPUs = 1344 GPU-hours
        # Physical fixed rate: $2.20/GPU-hr = $2956.80
        # Index mark price: $2.10/GPU-hr = $2822.40
        quote = self.publisher.compute_forward_hedge_quote(
            notional_hours=168,
            gpu_count=8,
            fixed_contract_rate=2.20,
            index_mark_price=2.10,
        )

        self.assertEqual(quote["total_gpu_hours"], 1344)
        self.assertEqual(quote["physical_contract_value_usd"], 2956.80)
        self.assertEqual(quote["floating_index_value_usd"], 2822.40)
        self.assertEqual(quote["basis_spread_usd"], 134.40)
        self.assertEqual(quote["recommended_hedge_side"], "SHORT_PERP")


if __name__ == "__main__":
    unittest.main()
