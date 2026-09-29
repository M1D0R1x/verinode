package auth

import (
	"testing"
	"time"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	h, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(h, "correct horse battery staple") {
		t.Error("valid password rejected")
	}
	if VerifyPassword(h, "wrong password") {
		t.Error("wrong password accepted")
	}
	if VerifyPassword("garbage", "x") {
		t.Error("malformed hash accepted")
	}
}

func TestJWTMintVerify(t *testing.T) {
	ts := NewTokenService("test-secret", time.Hour)
	tok, err := ts.Mint("usr_1", "a@b.com", RoleAdmin, "")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ts.Verify(tok)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Subject != "usr_1" || claims.Role != RoleAdmin {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestJWTTamperRejected(t *testing.T) {
	ts := NewTokenService("test-secret", time.Hour)
	tok, _ := ts.Mint("usr_1", "a@b.com", RoleTrader, "co_1")
	// Flip a char in the payload segment.
	tampered := tok[:len(tok)-4] + "AAAA"
	if _, err := ts.Verify(tampered); err == nil {
		t.Error("tampered token accepted")
	}
	// A token signed with a different secret must fail.
	other := NewTokenService("different-secret", time.Hour)
	tok2, _ := other.Mint("usr_1", "a@b.com", RoleTrader, "co_1")
	if _, err := ts.Verify(tok2); err == nil {
		t.Error("token signed with foreign secret accepted")
	}
}

func TestJWTExpiry(t *testing.T) {
	ts := NewTokenService("test-secret", time.Second)
	tok, _ := ts.Mint("usr_1", "a@b.com", RoleViewer, "co_1")
	if _, err := ts.Verify(tok); err != nil {
		t.Fatalf("fresh token should verify: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := ts.Verify(tok); err == nil {
		t.Error("expired token accepted")
	}
}

func TestRolePredicates(t *testing.T) {
	if !RoleSuperAdmin.IsPlatformStaff() || !RoleAdmin.IsPlatformStaff() {
		t.Error("super_admin/admin must be platform staff")
	}
	if RoleCompanyAdmin.IsPlatformStaff() || RoleTrader.IsPlatformStaff() || RoleViewer.IsPlatformStaff() {
		t.Error("company roles must NOT be platform staff")
	}
	if !RoleTrader.CanWriteTrading() || RoleViewer.CanWriteTrading() {
		t.Error("trader writes, viewer does not")
	}
	if !RoleCompanyAdmin.CanManageCompanyUsers() || RoleTrader.CanManageCompanyUsers() {
		t.Error("company_admin manages users, trader does not")
	}
}

func TestStoreLoginAndRegister(t *testing.T) {
	s := NewStore()
	// Seeded platform admin logs in.
	u, err := s.Authenticate("compliance@verinode.io", "verinode-admin")
	if err != nil || u.Role != RoleAdmin || u.CompanyID != "" {
		t.Fatalf("admin login failed: %v (%+v)", err, u)
	}
	// Wrong password rejected.
	if _, err := s.Authenticate("compliance@verinode.io", "nope"); err == nil {
		t.Error("wrong password accepted")
	}
	// Company self-registration yields a scoped company_admin, never platform staff.
	nu, err := s.RegisterCompany("Acme GPU Co", "founder@acme.example", "hunter2hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if nu.Role != RoleCompanyAdmin || nu.CompanyID == "" || nu.Role.IsPlatformStaff() {
		t.Errorf("registered company must be scoped company_admin, got %+v", nu)
	}
	// Duplicate email rejected.
	if _, err := s.RegisterCompany("Dupe", "founder@acme.example", "hunter2hunter2"); err == nil {
		t.Error("duplicate registration accepted")
	}
	// Short password rejected.
	if _, err := s.RegisterCompany("Weak", "weak@acme.example", "short"); err == nil {
		t.Error("short password accepted")
	}
}
