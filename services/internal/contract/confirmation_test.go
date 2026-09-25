package contract_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
)

func TestGenerateConfirmation_DeterministicDigestAndInvariants(t *testing.T) {
	c := &contract.ContractRecord{
		ID:              "c0a80101-0000-0000-0000-000000000001",
		GradeID:         "H100-SXM-8XNV",
		TemplateVersion: "v1.0.0-institutional",
		State:           contract.StateLive,
		CreatedAt:       time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC),
	}

	buyer := contract.ConfirmationParty{
		ID:           "buyer-001",
		LegalName:    "Anthropic Research SPV LLC",
		Jurisdiction: "Delaware, US",
		Role:         "buyer",
	}

	seller := contract.ConfirmationParty{
		ID:           "seller-001",
		LegalName:    "Crusoe Energy Infrastructure Corp",
		Jurisdiction: "Colorado, US",
		Role:         "seller",
	}

	doc1 := contract.GenerateConfirmation(c, buyer, seller, 168)
	doc2 := contract.GenerateConfirmation(c, buyer, seller, 168)

	// Determinism test
	if doc1.SHA256Checksum != doc2.SHA256Checksum {
		t.Fatalf("expected deterministic SHA-256 digest, got %s vs %s", doc1.SHA256Checksum, doc2.SHA256Checksum)
	}

	// Verify SHA-256 is 64 hex characters
	if len(doc1.SHA256Checksum) != 64 {
		t.Fatalf("expected 64 char hex hash, got len %d: %s", len(doc1.SHA256Checksum), doc1.SHA256Checksum)
	}

	// Verify Invariant 1 clause is present
	if !strings.Contains(doc1.DocumentContent, "INVARIANT 1 - PHYSICAL FORWARD EXCLUSION") {
		t.Fatalf("expected confirmation to contain Invariant 1 physical forward clause")
	}

	// Verify Invariant 4 clause is present
	if !strings.Contains(doc1.DocumentContent, "INVARIANT 4 - TELEMETRY ZERO-WORKLOAD PRIVACY") {
		t.Fatalf("expected confirmation to contain Invariant 4 privacy clause")
	}

	// Verify canonical trade_id is highlighted
	if !strings.Contains(doc1.DocumentContent, c.ID) {
		t.Fatalf("expected confirmation to embed canonical trade_id %s", c.ID)
	}

	// Verify that the hash in the document matches the recomputed hash
	parts := strings.Split(doc1.DocumentContent, "SHA-256 INTEGRITY DIGEST : ")
	if len(parts) != 2 {
		t.Fatalf("malformed document format")
	}
	preContent := parts[0]
	hasher := sha256.New()
	hasher.Write([]byte(preContent))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if expectedHash != doc1.SHA256Checksum {
		t.Fatalf("checksum mismatch: expected %s, got %s", expectedHash, doc1.SHA256Checksum)
	}
}
