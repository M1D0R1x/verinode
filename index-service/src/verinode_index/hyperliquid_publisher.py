"""
Hyperliquid HIP-3 Oracle Publisher & Physical Forward Hedging Adapter.
Publishes Verinode's canonical volume-weighted median index fixes to Hyperliquid L1.

Strict Invariant 5 Enforcement:
"The Index Engine must return insufficient_data: true when minimum contributor counts
or volume thresholds are unmet. Never interpolate or fabricate index prices."
"""

import time
from typing import Dict, Any, Optional
from .models import IndexFixResult


class HyperliquidOraclePublisher:
    """
    Adapter that reads canonical off-chain index calculations and publishes
    them as HIP-3 custom oracle market feeds on Hyperliquid.
    """

    def __init__(self, endpoint: str = "https://api.hyperliquid-testnet.xyz", oracle_name: str = "VERINODE-H100"):
        self.endpoint = endpoint
        self.oracle_name = oracle_name

    def format_hip3_oracle_envelope(self, fix: IndexFixResult) -> Dict[str, Any]:
        """
        Formats the HIP-3 oracle feed update.
        Rejects immediately if Invariant 5 is triggered (insufficient data).
        """
        if fix.insufficient_data:
            raise ValueError(
                f"Invariant 5 Violation Prevented: Refusing to publish to Hyperliquid HIP-3 oracle. "
                f"Reason: {fix.reason or 'Insufficient independent market depth'}"
            )

        if fix.value_usd is None or fix.value_usd <= 0:
            raise ValueError("Invalid price: benchmark value must be positive")

        now_ms = int(time.time() * 1000)

        return {
            "type": "hip3_oracle_update",
            "oracle_id": self.oracle_name,
            "series_id": fix.series_id,
            "mark_price": round(fix.value_usd, 4),
            "confidence_band": {
                "low": round(fix.confidence_interval_low, 4) if fix.confidence_interval_low else round(fix.value_usd * 0.95, 4),
                "high": round(fix.confidence_interval_high, 4) if fix.confidence_interval_high else round(fix.value_usd * 1.05, 4),
            },
            "contributor_count": fix.contributor_count,
            "total_notional_usd": fix.total_notional_usd,
            "methodology_version": fix.methodology_version,
            "timestamp_ms": now_ms,
            "status": "ACTIVE_FEED",
        }

    def compute_forward_hedge_quote(
        self,
        notional_hours: int,
        gpu_count: int,
        fixed_contract_rate: float,
        index_mark_price: float,
    ) -> Dict[str, Any]:
        """
        Calculates delta hedge requirement for an institutional infrastructure provider
        holding a physical capacity reservation looking to hedge on Hyperliquid.
        """
        total_gpu_hours = notional_hours * gpu_count
        physical_commitment_usd = total_gpu_hours * fixed_contract_rate
        index_floating_usd = total_gpu_hours * index_mark_price
        basis_spread_usd = physical_commitment_usd - index_floating_usd

        return {
            "total_gpu_hours": total_gpu_hours,
            "physical_contract_value_usd": round(physical_commitment_usd, 2),
            "floating_index_value_usd": round(index_floating_usd, 2),
            "basis_spread_usd": round(basis_spread_usd, 2),
            "recommended_hedge_side": "SHORT_PERP" if basis_spread_usd > 0 else "LONG_PERP",
            "recommended_hedge_size_contracts": round(total_gpu_hours, 2),
        }
