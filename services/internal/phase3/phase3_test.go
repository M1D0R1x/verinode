package phase3

import (
	"testing"
	"time"
)

func TestAdjustCredit_RewardsGoodHistory(t *testing.T) {
	d := AdjustCredit(CreditHistory{OnTimeSettlements: 19, LateSettlements: 1, CurrentLimitCents: 1_000_000})
	if d.DeltaCents <= 0 {
		t.Fatalf("expected positive delta for 95%% on-time, got %d (%s)", d.DeltaCents, d.Reason)
	}
	if d.NewLimitCents != 1_200_000 {
		t.Errorf("expected +20%% to 1,200,000, got %d", d.NewLimitCents)
	}
	if d.RequiresReview {
		t.Errorf("a bounded increase should not require review")
	}
}

func TestAdjustCredit_DefaultEscalates(t *testing.T) {
	d := AdjustCredit(CreditHistory{Defaults: 1, OnTimeSettlements: 10, CurrentLimitCents: 1_000_000})
	if !d.RequiresReview {
		t.Errorf("a default must require human review")
	}
	if d.DeltaCents >= 0 {
		t.Errorf("a default must decrease the limit, got delta %d", d.DeltaCents)
	}
}

func TestAdjustCredit_InsufficientHistory(t *testing.T) {
	d := AdjustCredit(CreditHistory{OnTimeSettlements: 2, CurrentLimitCents: 1_000_000})
	if d.DeltaCents != 0 {
		t.Errorf("expected no change with <5 settlements, got %d", d.DeltaCents)
	}
}

func TestRouteReplacement_PrefersSameGrade(t *testing.T) {
	now := time.Now().UTC()
	req := SubstitutionRequest{
		ContractID:    "c1",
		FailedGradeID: "H100-SXM-8XNV",
		WindowStart:   now.Add(time.Hour),
		WindowEnd:     now.Add(169 * time.Hour),
		OriginalPrice: 2956800,
		CureDeadline:  now.Add(48 * time.Hour),
		Candidates: []CandidateBlock{
			{BlockID: "b1", GradeID: "H100-SXM-8XNV", WindowStart: now, WindowEnd: now.Add(200 * time.Hour), PriceCents: 3000000},
			{BlockID: "b2", GradeID: "H100-SXM-8XNV", WindowStart: now, WindowEnd: now.Add(200 * time.Hour), PriceCents: 2900000},
		},
	}
	offer := RouteReplacement(req)
	if offer.Kind != "same_grade" || offer.BlockID != "b2" {
		t.Fatalf("expected cheapest same-grade b2, got %+v", offer)
	}
}

func TestRouteReplacement_SuperiorOnlyWithApproval(t *testing.T) {
	now := time.Now().UTC()
	base := SubstitutionRequest{
		FailedGradeID: "H100-SXM-8XNV",
		WindowStart:   now.Add(time.Hour),
		WindowEnd:     now.Add(169 * time.Hour),
		OriginalPrice: 2956800,
		CureDeadline:  now.Add(48 * time.Hour),
		Candidates: []CandidateBlock{
			{BlockID: "sup", GradeID: "H200-SXM-8XNV", WindowStart: now, WindowEnd: now.Add(200 * time.Hour), PriceCents: 4000000},
		},
	}
	// Without approval -> falls through to cash remedy.
	if got := RouteReplacement(base); got.Kind != "cash_remedy" {
		t.Errorf("expected cash_remedy without buyer approval, got %s", got.Kind)
	}
	// With approval -> superior grade.
	base.BuyerApprovesSup = true
	if got := RouteReplacement(base); got.Kind != "superior_grade" || got.OfferedGrade != "H200-SXM-8XNV" {
		t.Errorf("expected superior_grade H200, got %+v", got)
	}
}

func TestRouteReplacement_CashFloorWhenNothingFits(t *testing.T) {
	now := time.Now().UTC()
	req := SubstitutionRequest{
		FailedGradeID: "H100-SXM-8XNV",
		WindowStart:   now.Add(time.Hour),
		WindowEnd:     now.Add(169 * time.Hour),
		OriginalPrice: 2956800,
		CureDeadline:  now.Add(48 * time.Hour),
	}
	offer := RouteReplacement(req)
	if offer.Kind != "cash_remedy" || offer.CashRemedyCts != 2956800 {
		t.Fatalf("expected full cash remedy floor, got %+v", offer)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter()
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow("k", 5) {
			allowed++
		}
	}
	if allowed != 5 {
		t.Errorf("expected 5 allowed out of 10 for a 5/min bucket, got %d", allowed)
	}
}

func TestAPIKeyHashing(t *testing.T) {
	k, err := GenerateAPIKey([]string{"index:read"}, 60)
	if err != nil {
		t.Fatal(err)
	}
	if k.Secret == "" || k.Hash() == "" {
		t.Fatal("expected secret + hash")
	}
	if HashSecret(k.Secret) != k.Hash() {
		t.Errorf("hash mismatch: stored hash must equal HashSecret(secret)")
	}
	if len(k.Prefix) != 16 {
		t.Errorf("expected 16-char prefix, got %d", len(k.Prefix))
	}
}

func TestIdempotencyCache(t *testing.T) {
	c := NewIdempotencyCache(time.Minute)
	if _, _, ok := c.Get("x"); ok {
		t.Fatal("expected miss")
	}
	c.Put("x", 201, []byte(`{"ok":true}`))
	if body, status, ok := c.Get("x"); !ok || status != 201 || string(body) != `{"ok":true}` {
		t.Errorf("expected cached hit, got ok=%v status=%d body=%s", ok, status, body)
	}
}
