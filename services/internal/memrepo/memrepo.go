// Package memrepo provides in-memory implementations of the domain repositories,
// used when no DATABASE_URL is configured. This keeps the whole product functional
// end-to-end with zero infrastructure — the correct default for demos and for hosts
// where Docker/Postgres is unavailable. The pgx-backed repositories remain the
// production path; both satisfy the same method sets the API gateway calls.
//
// Data is seeded with a coherent demo world (participants -> inventory -> a live
// contract -> index series) so every screen shows real content immediately.
package memrepo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/inventory"
	"github.com/M1D0R1x/verinode/services/internal/participant"
)

func uid() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:16])
}

// Demo identity ids shared across seeded repos so relationships line up.
const (
	DemoBuyerID   = "00000000-0000-0000-0000-000000000001"
	DemoSellerID  = "00000000-0000-0000-0000-000000000002"
	DemoBuyer2ID  = "00000000-0000-0000-0000-000000000003"
	DemoContractID = "c37610f3-4aea-414f-868c-5937e97e822d"
	DemoBlockID   = "b1000000-0000-0000-0000-000000000001"
)

// ---------------------------------------------------------------------------
// Participant
// ---------------------------------------------------------------------------

type ParticipantRepo struct {
	mu sync.RWMutex
	m  map[string]*participant.Participant
}

func NewParticipantRepo() *ParticipantRepo {
	r := &ParticipantRepo{m: make(map[string]*participant.Participant)}
	now := time.Now().UTC()
	limit := int64(500_000_000)
	seed := []*participant.Participant{
		{ID: DemoBuyerID, LegalName: "Anthropic Compute SPV", Jurisdiction: "US", Role: "buyer", KYCStatus: "approved", CreditLimitCents: &limit, CreatedAt: now, UpdatedAt: now},
		{ID: DemoSellerID, LegalName: "Lambda Labs", Jurisdiction: "US", Role: "seller", KYCStatus: "approved", CreatedAt: now, UpdatedAt: now},
		{ID: DemoBuyer2ID, LegalName: "Mistral Frontier Ltd", Jurisdiction: "FR", Role: "buyer", KYCStatus: "pending", CreatedAt: now, UpdatedAt: now},
	}
	for _, p := range seed {
		r.m[p.ID] = p
	}
	return r
}

func (r *ParticipantRepo) Create(ctx context.Context, p *participant.Participant) error {
	if p.LegalName == "" || p.Jurisdiction == "" || p.Role == "" {
		return participant.ErrInvalidParticipant
	}
	if p.KYCStatus == "" {
		p.KYCStatus = "pending"
	}
	p.ID = uid()
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	r.mu.Lock()
	r.m[p.ID] = p
	r.mu.Unlock()
	return nil
}

func (r *ParticipantRepo) GetByID(ctx context.Context, id string) (*participant.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.m[id]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, participant.ErrParticipantNotFound
}

func (r *ParticipantRepo) List(ctx context.Context) ([]participant.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]participant.Participant, 0, len(r.m))
	for _, p := range r.m {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *ParticipantRepo) UpdateKYC(ctx context.Context, id, status string, creditLimitCents *int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.m[id]
	if !ok {
		return participant.ErrParticipantNotFound
	}
	p.KYCStatus = status
	if creditLimitCents != nil {
		p.CreditLimitCents = creditLimitCents
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ---------------------------------------------------------------------------
// Inventory
// ---------------------------------------------------------------------------

type InventoryRepo struct {
	mu sync.RWMutex
	m  map[string]*inventory.InventoryBlock
}

func NewInventoryRepo() *InventoryRepo {
	r := &InventoryRepo{m: make(map[string]*inventory.InventoryBlock)}
	now := time.Now().UTC()
	seed := []*inventory.InventoryBlock{
		{ID: DemoBlockID, SellerID: DemoSellerID, GradeID: "H100-SXM-8XNV", RegionBucket: "US-East", WindowStart: now.Add(24 * time.Hour), WindowEnd: now.Add(192 * time.Hour), Status: "available", CreatedAt: now, UpdatedAt: now},
		{ID: uid(), SellerID: DemoSellerID, GradeID: "H100-SXM-8XNV", RegionBucket: "US-West", WindowStart: now.Add(48 * time.Hour), WindowEnd: now.Add(216 * time.Hour), Status: "available", CreatedAt: now, UpdatedAt: now},
		{ID: uid(), SellerID: DemoSellerID, GradeID: "H100-SXM-8XNV", RegionBucket: "EU-Central", WindowStart: now.Add(72 * time.Hour), WindowEnd: now.Add(240 * time.Hour), Status: "available", CreatedAt: now, UpdatedAt: now},
	}
	for _, b := range seed {
		r.m[b.ID] = b
	}
	return r
}

func (r *InventoryRepo) Create(ctx context.Context, b *inventory.InventoryBlock) error {
	if b.SellerID == "" || b.GradeID == "" || b.RegionBucket == "" {
		return inventory.ErrInvalidBlock
	}
	if !b.WindowEnd.After(b.WindowStart) {
		return inventory.ErrInvalidWindow
	}
	if b.Status == "" {
		b.Status = "available"
	}
	b.ID = uid()
	b.CreatedAt = time.Now().UTC()
	b.UpdatedAt = b.CreatedAt
	r.mu.Lock()
	r.m[b.ID] = b
	r.mu.Unlock()
	return nil
}

func (r *InventoryRepo) GetByID(ctx context.Context, id string) (*inventory.InventoryBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if b, ok := r.m[id]; ok {
		cp := *b
		return &cp, nil
	}
	return nil, inventory.ErrBlockNotFound
}

func (r *InventoryRepo) FindAvailableBlocks(ctx context.Context, gradeID, regionBucket string, start, end time.Time) ([]inventory.InventoryBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []inventory.InventoryBlock
	for _, b := range r.m {
		if b.Status == "available" && b.GradeID == gradeID && (regionBucket == "" || b.RegionBucket == regionBucket) {
			out = append(out, *b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WindowStart.Before(out[j].WindowStart) })
	return out, nil
}

var errNotImplemented = errors.New("not implemented in memory store")
