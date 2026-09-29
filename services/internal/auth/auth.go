// Package auth implements Verinode's authentication and role-based access control.
//
// Role model (see docs/07-auth-and-rbac.md for the justification):
//
//	super_admin   platform owner — everything, incl. user/role management
//	admin         platform ops/compliance — the /admin console (participant approval,
//	              surveillance, claims, cross-company audit)
//	company_admin a registered company's owner — manages that company's users and sees
//	              ONLY that company's RFQs/contracts/claims (scoped org view)
//	trader        company member — create RFQs, respond to quotes, within their company
//	viewer        company member — read-only, within their company
//
// The platform admin console is restricted to super_admin + admin. Registered companies
// never see the platform console — they get scoped self-service — because surveillance,
// participant approval and cross-company audit are platform trust-domain functions and
// exposing them would leak competitors' data and break the "independent" claim.
//
// Crypto is stdlib-only (no external deps): PBKDF2-HMAC-SHA256 password hashing and
// HMAC-SHA256 (HS256) compact JWTs. In production this is swapped for the OIDC/SAML
// vendor named in the architecture docs; the middleware boundary stays identical.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Role string

const (
	RoleSuperAdmin   Role = "super_admin"
	RoleAdmin        Role = "admin"
	RoleCompanyAdmin Role = "company_admin"
	RoleTrader       Role = "trader"
	RoleViewer       Role = "viewer"
)

var validRoles = map[Role]bool{
	RoleSuperAdmin: true, RoleAdmin: true, RoleCompanyAdmin: true, RoleTrader: true, RoleViewer: true,
}

func (r Role) Valid() bool { return validRoles[r] }

// IsPlatformStaff reports whether the role may access the platform admin console.
func (r Role) IsPlatformStaff() bool { return r == RoleSuperAdmin || r == RoleAdmin }

// CanWriteTrading reports whether the role may create RFQs / respond to quotes.
func (r Role) CanWriteTrading() bool {
	return r == RoleSuperAdmin || r == RoleCompanyAdmin || r == RoleTrader
}

// CanManageCompanyUsers reports whether the role may manage its company's users.
func (r Role) CanManageCompanyUsers() bool {
	return r == RoleSuperAdmin || r == RoleCompanyAdmin
}

var (
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrForbidden      = errors.New("insufficient role for this action")
	ErrBadCredentials = errors.New("invalid email or password")
)

// Claims is the JWT payload for an authenticated principal.
type Claims struct {
	Subject   string `json:"sub"`        // user id
	Email     string `json:"email"`
	Role      Role   `json:"role"`
	CompanyID string `json:"company_id"` // "" for platform staff
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// -----------------------------------------------------------------------------
// PBKDF2-HMAC-SHA256 password hashing (stdlib only)
// -----------------------------------------------------------------------------

const (
	pbkdf2Iters = 120_000
	pbkdf2KeyLen = 32
	saltLen      = 16
)

func pbkdf2(password, salt []byte, iters, keyLen int) []byte {
	h := sha256.New
	hashLen := sha256.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen
	var dk []byte
	buf := make([]byte, 4)
	for block := 1; block <= numBlocks; block++ {
		prf := hmac.New(h, password)
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf, uint32(block))
		prf.Write(buf)
		u := prf.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for n := 2; n <= iters; n++ {
			prf = hmac.New(h, password)
			prf.Write(u)
			u = prf.Sum(nil)
			for x := range t {
				t[x] ^= u[x]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

// HashPassword returns an encoded "pbkdf2_sha256$iters$salt$hash" string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2([]byte(password), salt, pbkdf2Iters, pbkdf2KeyLen)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		pbkdf2Iters,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

// VerifyPassword checks a plaintext password against an encoded hash in constant time.
func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	var iters int
	if _, err := fmt.Sscanf(parts[1], "%d", &iters); err != nil || iters <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got := pbkdf2([]byte(password), salt, iters, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// -----------------------------------------------------------------------------
// HS256 compact JWT (stdlib only)
// -----------------------------------------------------------------------------

// TokenService mints and verifies signed session tokens.
type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	if secret == "" {
		secret = "verinode-dev-insecure-secret-change-me"
	}
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// Mint creates a signed HS256 JWT for the principal.
func (t *TokenService) Mint(userID, email string, role Role, companyID string) (string, error) {
	now := time.Now()
	claims := Claims{
		Subject:   userID,
		Email:     email,
		Role:      role,
		CompanyID: companyID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(t.ttl).Unix(),
	}
	header := b64([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadJSON, _ := json.Marshal(claims)
	payload := b64(payloadJSON)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(signingInput))
	sig := b64(mac.Sum(nil))
	return signingInput + "." + sig, nil
}

// Verify validates a token's signature and expiry and returns its claims.
func (t *TokenService) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(signingInput))
	expectedSig := b64(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[2])) != 1 {
		return nil, ErrInvalidToken
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, ErrInvalidToken
	}
	if !claims.Role.Valid() {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}
