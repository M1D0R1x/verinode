"""
Verinode Canonical Index Artifact Publisher
Formats canonical signed index fix publications.
"""

import json
from dataclasses import asdict
from typing import Dict, Any
from .models import IndexFixResult


class IndexPublisher:
    @staticmethod
    def format_publication_artifact(result: IndexFixResult) -> Dict[str, Any]:
        """Produces canonical JSON dictionary artifact for publication."""
        payload = asdict(result)
        return {
            "schema_version": "v1.0.0",
            "series": payload["series_id"],
            "insufficient_data": payload["insufficient_data"],
            "benchmark_price_usd": payload["value_usd"],
            "unit": payload["unit"],
            "dispersion": {
                "p25": payload["confidence_interval_low"],
                "p75": payload["confidence_interval_high"],
            },
            "metrics": {
                "contributor_count": payload["contributor_count"],
                "observation_count": payload["observation_count"],
                "total_notional_usd": payload["total_notional_usd"],
            },
            "published_at": payload["publish_time"],
            "methodology": payload["methodology_version"],
            "status_reason": payload["reason"],
            "canonical_format": "JSON_RFC8259",
        }
