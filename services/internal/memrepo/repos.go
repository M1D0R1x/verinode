package memrepo

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/claims"
	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/marketdata"
	"github.com/M1D0R1x/verinode/services/internal/rfq"
)

// ---------------------------------------------------------------------------
// RFQ + Quotes
// ---------------------------------------------------------------------------

type RFQRepo struct {
	mu     sync.RWMutex
	rfqs   map[string]*rfq.RFQ
	quotes map[string]*rfq.Quote
}

func NewRFQRepo() *RFQRepo {
	return &RFQRepo{rfqs: make(map[string]*rfq.RFQ), quotes: make(map[string]*rfq.Quote)}
}

func (r *RFQRepo) CreateRFQ(ctx context.Context, req *rfq.RFQ) error {
	if req.GradeID == "" || req.RegionBucket == "" {
		return rfq.ErrInvalidRFQ
	}
	if !req.WindowEnd.After(req.WindowStart) {
		return rfq.ErrInvalidRFQ
	}
	req.ID = uid()
	if req.Status == "" {
		req.Status = "open"
	}
	req.CreatedAt = time.Now().UTC()
	r.mu.Lock()
	r.rfqs[req.ID] = req
	r.mu.Unlock()
	return nil
}

func (r *RFQRepo) GetRFQByID(ctx context.Context, id string) (*rfq.RFQ, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if v, ok := r.rfqs[id]; ok {
		cp := *v
		return &cp, nil
	}
	return nil, rfq.ErrRFQNotFound
}

func (r *RFQRepo) CreateQuote(ctx context.Context, q *rfq.Quote) error {
	if q.RFQID == "" || q.PriceCents <= 0 {
		return rfq.ErrInvalidQuote
	}
	q.ID = uid()
	if q.Status == "" {
		q.Status = "active"
	}
	if q.Currency == "" {
		q.Currency = "USD"
	}
	q.CreatedAt = time.Now().UTC()
	r.mu.Lock()
	r.quotes[q.ID] = q
	if rf, ok := r.rfqs[q.RFQID]; ok {
		rf.Status = "quoted"
	}
	r.mu.Unlock()
	return nil
}

func (r *RFQRepo) GetQuotesByRFQ(ctx context.Context, rfqID string) ([]rfq.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []rfq.Quote
	for _, q := range r.quotes {
		if q.RFQID == rfqID {
			out = append(out, *q)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *RFQRepo) AcceptQuote(ctx context.Context, quoteID, buyerID string) (*rfq.AcceptedQuoteDetails, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q, ok := r.quotes[quoteID]
	if !ok {
		return nil, rfq.ErrQuoteNotFound
	}
	if q.Status != "active" {
		return nil, rfq.ErrQuoteAlreadyTaken
	}
	if !q.ExpiresAt.IsZero() && time.Now().After(q.ExpiresAt) {
		return nil, rfq.ErrQuoteExpired
	}
	rf, ok := r.rfqs[q.RFQID]
	if !ok {
		return nil, rfq.ErrRFQNotFound
	}
	if rf.BuyerID != "" && buyerID != "" && rf.BuyerID != buyerID {
		return nil, rfq.ErrUnauthorizedBuyer
	}
	q.Status = "accepted"
	rf.Status = "accepted"
	return &rfq.AcceptedQuoteDetails{
		QuoteID: q.ID, RFQID: rf.ID, BuyerID: buyerID, SellerID: q.SellerID,
		BlockID: q.BlockID, GradeID: rf.GradeID, PriceCents: q.PriceCents, Currency: q.Currency,
	}, nil
}

// ---------------------------------------------------------------------------
// Contract + events (append-only, state-machine validated)
// ---------------------------------------------------------------------------

type ContractRepo struct {
	mu       sync.RWMutex
	m        map[string]*contract.ContractRecord
	events   map[string][]contract.ContractEventRecord
	idemSeen map[string]bool // contractID|idempotencyKey
}

func NewContractRepo() *ContractRepo {
	r := &ContractRepo{
		m:        make(map[string]*contract.ContractRecord),
		events:   make(map[string][]contract.ContractEventRecord),
		idemSeen: make(map[string]bool),
	}
	// Seed one live demo contract so admin/contract screens show content.
	now := time.Now().UTC()
	rfqID, quoteID := "rfq-demo", "quote-demo"
	hash := "4adec53cc851db49c10c317c5fc507e31a3f166844e47c660e28a4da74abd775"
	ref := "s3://verinode-confirmations/" + DemoContractID + ".pdf"
	r.m[DemoContractID] = &contract.ContractRecord{
		ID: DemoContractID, RFQID: &rfqID, QuoteID: &quoteID,
		BuyerID: DemoBuyerID, SellerID: DemoSellerID, GradeID: "H100-SXM-8XNV",
		TemplateVersion: "v1.0.0-institutional", State: contract.StateLive,
		SignedPDFHash: &hash, SignedPDFRef: &ref, CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now,
	}
	r.events[DemoContractID] = []contract.ContractEventRecord{
		{ID: uid(), ContractID: DemoContractID, PriorState: "contract_pending", NewState: "funded_secured", Actor: "ledger", Reason: "deposit confirmed", IdempotencyKey: "seed-1", CreatedAt: now.Add(-60 * time.Hour)},
		{ID: uid(), ContractID: DemoContractID, PriorState: "funded_secured", NewState: "scheduled", Actor: "inventory", Reason: "capacity reserved", IdempotencyKey: "seed-2", CreatedAt: now.Add(-48 * time.Hour)},
		{ID: uid(), ContractID: DemoContractID, PriorState: "scheduled", NewState: "delivery_test", Actor: "telemetry", Reason: "canary window opened", IdempotencyKey: "seed-3", CreatedAt: now.Add(-26 * time.Hour)},
		{ID: uid(), ContractID: DemoContractID, PriorState: "delivery_test", NewState: "live", Actor: "delivery", Reason: "canary passed 428 GB/s", IdempotencyKey: "seed-4", CreatedAt: now.Add(-24 * time.Hour)},
	}
	return r
}

func (r *ContractRepo) CreateContract(ctx context.Context, c *contract.ContractRecord) error {
	if c.ID == "" {
		c.ID = uid()
	}
	if c.State == "" {
		c.State = contract.StateContractPending
	}
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt
	r.mu.Lock()
	r.m[c.ID] = c
	r.mu.Unlock()
	return nil
}

func (r *ContractRepo) GetByID(ctx context.Context, tradeID string) (*contract.ContractRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.m[tradeID]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, contract.ErrContractNotFound
}

func (r *ContractRepo) AdvanceContractState(ctx context.Context, tradeID string, nextState contract.State, evt contract.TransitionEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.m[tradeID]
	if !ok {
		return contract.ErrContractNotFound
	}
	if evt.IdempotencyKey != "" {
		k := tradeID + "|" + evt.IdempotencyKey
		if r.idemSeen[k] {
			return nil // idempotent replay
		}
		r.idemSeen[k] = true
	}
	if err := contract.ValidateTransition(c.State, nextState, evt); err != nil {
		return err
	}
	prior := c.State
	c.State = nextState
	c.UpdatedAt = time.Now().UTC()
	var evHash *string
	if evt.EvidenceHash != "" {
		evHash = &evt.EvidenceHash
	}
	r.events[tradeID] = append(r.events[tradeID], contract.ContractEventRecord{
		ID: uid(), ContractID: tradeID, PriorState: string(prior), NewState: string(nextState),
		Actor: evt.Actor, Reason: evt.Reason, IdempotencyKey: evt.IdempotencyKey,
		EvidenceHash: evHash, AuthorizationDecision: evt.AuthorizationDecision, CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (r *ContractRepo) ListContracts(ctx context.Context) ([]contract.ContractRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]contract.ContractRecord, 0, len(r.m))
	for _, c := range r.m {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *ContractRepo) GetContractEvents(ctx context.Context, tradeID string) ([]contract.ContractEventRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	evs := append([]contract.ContractEventRecord(nil), r.events[tradeID]...)
	sort.Slice(evs, func(i, j int) bool { return evs[i].CreatedAt.Before(evs[j].CreatedAt) })
	return evs, nil
}

func (r *ContractRepo) ListAllEvents(ctx context.Context, limit int) ([]contract.ContractEventRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var all []contract.ContractEventRecord
	for _, evs := range r.events {
		all = append(all, evs...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *ContractRepo) UpdateSignedPDFHash(ctx context.Context, tradeID, hash, ref string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.m[tradeID]
	if !ok {
		return contract.ErrContractNotFound
	}
	c.SignedPDFHash = &hash
	c.SignedPDFRef = &ref
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// ---------------------------------------------------------------------------
// Claims
// ---------------------------------------------------------------------------

type ClaimsRepo struct {
	mu sync.RWMutex
	m  map[string]*claims.Claim
}

func NewClaimsRepo() *ClaimsRepo {
	return &ClaimsRepo{m: make(map[string]*claims.Claim)}
}

func (r *ClaimsRepo) Create(ctx context.Context, c *claims.Claim) error {
	if c.ContractID == "" || c.Type == "" {
		return claims.ErrInvalidClaim
	}
	c.ID = uid()
	if c.State == "" {
		c.State = "claim_open"
	}
	c.CreatedAt = time.Now().UTC()
	r.mu.Lock()
	r.m[c.ID] = c
	r.mu.Unlock()
	return nil
}

func (r *ClaimsRepo) List(ctx context.Context, stateFilter string) ([]claims.Claim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []claims.Claim
	for _, c := range r.m {
		if stateFilter == "" || c.State == stateFilter {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *ClaimsRepo) Resolve(ctx context.Context, claimID, targetState string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.m[claimID]
	if !ok {
		return claims.ErrClaimNotFound
	}
	c.State = targetState
	now := time.Now().UTC()
	c.ResolvedAt = &now
	return nil
}

// ---------------------------------------------------------------------------
// Market data / index / surveillance
// ---------------------------------------------------------------------------

type MarketRepo struct {
	mu       sync.RWMutex
	series   map[string]*marketdata.IndexSeries
	obs      map[string][]marketdata.IndexObservation
	contribs map[string][]marketdata.MarketContribution
	flags    map[string]*marketdata.SurveillanceFlag
}

const DemoSeriesID = "H100-SXM-8XNV-US-WEEK-DEDICATED-USD"

func NewMarketRepo() *MarketRepo {
	r := &MarketRepo{
		series:   make(map[string]*marketdata.IndexSeries),
		obs:      make(map[string][]marketdata.IndexObservation),
		contribs: make(map[string][]marketdata.MarketContribution),
		flags:    make(map[string]*marketdata.SurveillanceFlag),
	}
	now := time.Now().UTC()
	r.series[DemoSeriesID] = &marketdata.IndexSeries{
		ID: DemoSeriesID, GPUModel: "H100", Form: "SXM", Topology: "8xNVLink",
		RegionBucket: "US", Tenor: "168h", Tenancy: "dedicated", Currency: "USD",
		MethodologyVersion: "v1.0.0-vw-median", MinContributors: 3, MinNotionalUSD: 50000, MaxContributorWeight: 0.35,
		CreatedAt: now.Add(-30 * 24 * time.Hour),
	}
	// Seed a week of published observations so the index chart has content.
	base := 2.05
	for i := 6; i >= 0; i-- {
		v := base + 0.03*float64(6-i)
		lo, hi := v-0.08, v+0.08
		sig := "ed25519:seed"
		r.obs[DemoSeriesID] = append(r.obs[DemoSeriesID], marketdata.IndexObservation{
			ID: uid(), SeriesID: DemoSeriesID, Value: &v, Unit: "USD_PER_GPU_HOUR",
			ObservationWindowStart: now.Add(time.Duration(-(i + 1)*24) * time.Hour),
			ObservationWindowEnd:   now.Add(time.Duration(-i*24) * time.Hour),
			PublishTime:            now.Add(time.Duration(-i*24) * time.Hour),
			SequenceNumber:         int64(7 - i), ContributorCount: 4, ObservationCount: 12, TotalNotionalUSD: 220000,
			ConfidenceIntervalLow: &lo, ConfidenceIntervalHigh: &hi, InsufficientData: false, Signature: &sig,
		})
	}
	// One pending surveillance flag so the desk isn't empty.
	r.flags[uid()] = &marketdata.SurveillanceFlag{
		ID: uid(), SubjectType: "contribution", SubjectID: "contrib-demo", FlagType: "related_party",
		Severity: "warning", Details: map[string]interface{}{"note": "two contributions share a beneficial owner"},
		Status: "pending", CreatedAt: now.Add(-4 * time.Hour),
	}
	return r
}

func (r *MarketRepo) GetSeries(ctx context.Context, id string) (*marketdata.IndexSeries, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if s, ok := r.series[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, marketdata.ErrSeriesNotFound
}

func (r *MarketRepo) ListSeries(ctx context.Context) ([]marketdata.IndexSeries, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]marketdata.IndexSeries, 0, len(r.series))
	for _, s := range r.series {
		out = append(out, *s)
	}
	return out, nil
}

func (r *MarketRepo) UpsertSeries(ctx context.Context, s *marketdata.IndexSeries) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *s
	r.series[s.ID] = &cp
	return nil
}

func (r *MarketRepo) GetLatestObservation(ctx context.Context, seriesID string) (*marketdata.IndexObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.obs[seriesID]
	if len(list) == 0 {
		return nil, nil
	}
	latest := list[0]
	for _, o := range list {
		if o.SequenceNumber > latest.SequenceNumber {
			latest = o
		}
	}
	cp := latest
	return &cp, nil
}

func (r *MarketRepo) ListObservations(ctx context.Context, seriesID string, limit int) ([]marketdata.IndexObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := append([]marketdata.IndexObservation(nil), r.obs[seriesID]...)
	sort.Slice(list, func(i, j int) bool { return list[i].SequenceNumber > list[j].SequenceNumber })
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *MarketRepo) RecordObservation(ctx context.Context, obs *marketdata.IndexObservation) error {
	if obs.ID == "" {
		obs.ID = uid()
	}
	if obs.CreatedAt.IsZero() {
		obs.CreatedAt = time.Now().UTC()
	}
	r.mu.Lock()
	r.obs[obs.SeriesID] = append(r.obs[obs.SeriesID], *obs)
	r.mu.Unlock()
	return nil
}

func (r *MarketRepo) RecordContribution(ctx context.Context, c *marketdata.MarketContribution) error {
	if c.ID == "" {
		c.ID = uid()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	r.mu.Lock()
	r.contribs[c.SeriesID] = append(r.contribs[c.SeriesID], *c)
	r.mu.Unlock()
	return nil
}

func (r *MarketRepo) ListContributions(ctx context.Context, seriesID string, since time.Time) ([]marketdata.MarketContribution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []marketdata.MarketContribution
	for _, c := range r.contribs[seriesID] {
		if c.CreatedAt.After(since) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *MarketRepo) ListSurveillanceFlags(ctx context.Context, status string) ([]marketdata.SurveillanceFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []marketdata.SurveillanceFlag
	for _, f := range r.flags {
		if status == "" || f.Status == status {
			out = append(out, *f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *MarketRepo) CreateSurveillanceFlag(ctx context.Context, f *marketdata.SurveillanceFlag) error {
	if f.ID == "" {
		f.ID = uid()
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	r.mu.Lock()
	r.flags[f.ID] = f
	r.mu.Unlock()
	return nil
}

func (r *MarketRepo) ReviewSurveillanceFlag(ctx context.Context, id, reviewer, resolution, newStatus string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.flags[id]
	if !ok {
		return marketdata.ErrFlagNotFound
	}
	f.ReviewedBy = &reviewer
	f.Resolution = &resolution
	f.Status = newStatus
	return nil
}
