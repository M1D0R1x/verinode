package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

// User is a platform or company principal.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	CompanyID    string    `json:"company_id,omitempty"`
	CompanyName  string    `json:"company_name,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store is a thread-safe user store. In-memory here; the same interface backs a
// Postgres implementation in production without changing the HTTP layer.
type Store struct {
	mu       sync.RWMutex
	byEmail  map[string]*User
	byID     map[string]*User
}

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

// NewStore builds a store seeded with a platform super-admin and a demo company so the
// product is usable immediately. Credentials are logged once at startup in dev.
func NewStore() *Store {
	s := &Store{byEmail: make(map[string]*User), byID: make(map[string]*User)}

	seed := func(email, pw string, role Role, companyID, companyName string) {
		hash, _ := HashPassword(pw)
		u := &User{
			ID:           newID("usr"),
			Email:        strings.ToLower(email),
			Role:         role,
			CompanyID:    companyID,
			CompanyName:  companyName,
			PasswordHash: hash,
			CreatedAt:    time.Now().UTC(),
		}
		s.byEmail[u.Email] = u
		s.byID[u.ID] = u
	}

	// Platform staff (no company scope).
	seed("root@verinode.io", "verinode-super-admin", RoleSuperAdmin, "", "")
	seed("compliance@verinode.io", "verinode-admin", RoleAdmin, "", "")
	// Demo company principals (company-scoped).
	seed("admin@anthropic-spv.example", "demo-company-admin", RoleCompanyAdmin, "co_demo_buyer", "Anthropic Compute SPV")
	seed("trader@anthropic-spv.example", "demo-trader", RoleTrader, "co_demo_buyer", "Anthropic Compute SPV")
	seed("ops@lambda-labs.example", "demo-seller-admin", RoleCompanyAdmin, "co_demo_seller", "Lambda Labs")

	return s
}

// SeededCredentials returns the demo logins for surfacing on the login screen (dev only).
func (s *Store) SeededCredentials() []map[string]string {
	return []map[string]string{
		{"role": "super_admin", "email": "root@verinode.io", "password": "verinode-super-admin"},
		{"role": "admin", "email": "compliance@verinode.io", "password": "verinode-admin"},
		{"role": "company_admin (buyer)", "email": "admin@anthropic-spv.example", "password": "demo-company-admin"},
		{"role": "trader (buyer)", "email": "trader@anthropic-spv.example", "password": "demo-trader"},
		{"role": "company_admin (seller)", "email": "ops@lambda-labs.example", "password": "demo-seller-admin"},
	}
}

// Authenticate verifies credentials and returns the user.
func (s *Store) Authenticate(email, password string) (*User, error) {
	s.mu.RLock()
	u, ok := s.byEmail[strings.ToLower(strings.TrimSpace(email))]
	s.mu.RUnlock()
	if !ok || !VerifyPassword(u.PasswordHash, password) {
		return nil, ErrBadCredentials
	}
	return u, nil
}

// RegisterCompany creates a new company and its company_admin principal. Companies
// self-register into the scoped org view; they never become platform staff.
func (s *Store) RegisterCompany(companyName, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if companyName == "" || email == "" || len(password) < 8 {
		return nil, ErrBadCredentials
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[email]; exists {
		return nil, ErrBadCredentials
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &User{
		ID:           newID("usr"),
		Email:        email,
		Role:         RoleCompanyAdmin,
		CompanyID:    newID("co"),
		CompanyName:  companyName,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	s.byEmail[email] = u
	s.byID[u.ID] = u
	return u, nil
}

// GetByID returns a user by id.
func (s *Store) GetByID(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	return u, ok
}

// -----------------------------------------------------------------------------
// HTTP middleware & context
// -----------------------------------------------------------------------------

type ctxKey int

const claimsKey ctxKey = 0

// FromContext extracts the authenticated claims placed by RequireAuth.
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// bearer extracts the token from Authorization or a session cookie.
func bearer(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie("vn_session"); err == nil {
		return c.Value
	}
	return ""
}

// Authenticate returns the verified claims for a request, or an error.
func (t *TokenService) Authenticate(r *http.Request) (*Claims, error) {
	tok := bearer(r)
	if tok == "" {
		return nil, ErrInvalidToken
	}
	return t.Verify(tok)
}

// WithClaims stores claims on the request context.
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}
