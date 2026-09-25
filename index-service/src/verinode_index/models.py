"""
Verinode Institutional Index & Market Data Models
Defines immutable observation contracts and index fix results.
"""

from dataclasses import dataclass
from datetime import datetime
from enum import IntEnum
from typing import Optional


class InputTier(IntEnum):
    """Hierarchy of inputs per Verinode Index Methodology §4"""
    COMPLETED_TRADE = 1     # Highest weight (executed physical delivery)
    MATCHED_TRADE = 2       # Signed contract pending delivery
    FIRM_TWO_SIDED_QUOTE = 3 # Executable bid/ask spread
    FIRM_ONE_SIDED_QUOTE = 4 # Firm supplier quote


@dataclass(frozen=True)
class MarketObservation:
    """An individual atomic market price point from vetted participants."""
    contributor_id: str
    hourly_price_usd: float
    duration_hours: int
    tier: InputTier
    timestamp: datetime
    trade_id: Optional[str] = None

    @property
    def notional_usd(self) -> float:
        return self.hourly_price_usd * self.duration_hours


@dataclass
class IndexFixResult:
    """The authoritative index benchmark fix for a designated delivery grade."""
    series_id: str
    unit: str
    insufficient_data: bool
    contributor_count: int
    observation_count: int
    total_notional_usd: float
    publish_time: str
    methodology_version: str = "v1.0.0-institutional"
    value_usd: Optional[float] = None
    confidence_interval_low: Optional[float] = None
    confidence_interval_high: Optional[float] = None
    reason: Optional[str] = None
    signature: Optional[str] = None
