package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type ConfirmationParty struct {
	ID           string `json:"id"`
	LegalName    string `json:"legal_name"`
	Jurisdiction string `json:"jurisdiction"`
	Role         string `json:"role"`
}

type ConfirmationDocument struct {
	TradeID         string            `json:"trade_id"`
	ContractDate    time.Time         `json:"contract_date"`
	Buyer           ConfirmationParty `json:"buyer"`
	Seller          ConfirmationParty `json:"seller"`
	GradeID         string            `json:"grade_id"`
	TemplateVersion string            `json:"template_version"`
	DurationHours   int               `json:"duration_hours"`
	State           string            `json:"state"`
	SHA256Checksum  string            `json:"sha256_checksum"`
	DocumentContent string            `json:"document_content"`
}

// GenerateConfirmation builds the institutional Physical Forward Reservation Confirmation
// and computes the canonical SHA-256 hash across the standardized document envelope.
func GenerateConfirmation(c *ContractRecord, buyer, seller ConfirmationParty, durationHours int) *ConfirmationDocument {
	if durationHours <= 0 {
		durationHours = 168 // Default benchmark period (7 days)
	}

	doc := strings.Builder{}
	doc.WriteString("================================================================================\n")
	doc.WriteString("               VERINODE INSTITUTIONAL GPU CAPACITY RESERVATION                  \n")
	doc.WriteString("                     MASTER PHYSICAL FORWARD CONFIRMATION                       \n")
	doc.WriteString("================================================================================\n\n")

	doc.WriteString(fmt.Sprintf("TRANSACTION ID (TRADE_ID) : %s\n", c.ID))
	doc.WriteString(fmt.Sprintf("CONFIRMATION DATE         : %s\n", c.CreatedAt.UTC().Format(time.RFC3339)))
	doc.WriteString(fmt.Sprintf("MASTER TEMPLATE VERSION   : %s\n", c.TemplateVersion))
	doc.WriteString(fmt.Sprintf("CURRENT CONTRACT STATE    : %s\n\n", string(c.State)))

	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString("1. CONTRACTING PARTIES (BILATERAL COUNTERPARTIES)\n")
	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString(fmt.Sprintf("BUYER ENTITY:\n  Legal Name   : %s\n  Entity ID    : %s\n  Jurisdiction : %s\n\n",
		buyer.LegalName, buyer.ID, buyer.Jurisdiction))
	doc.WriteString(fmt.Sprintf("SELLER ENTITY:\n  Legal Name   : %s\n  Entity ID    : %s\n  Jurisdiction : %s\n\n",
		seller.LegalName, seller.ID, seller.Jurisdiction))

	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString("2. PHYSICAL COMMODITY SPECIFICATION & BENCHMARK GRADE\n")
	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString(fmt.Sprintf("BENCHMARK GRADE       : %s\n", c.GradeID))
	doc.WriteString("ACCELERATOR TOPOLOGY  : 8x NVIDIA H100 SXM5 (80GB HBM3 each, 640GB aggregate)\n")
	doc.WriteString("INTERCONNECT BUS      : SXM5 / HGX 8-Way NVLink 4.0 / NVSwitch (900 GB/s bidirectional)\n")
	doc.WriteString("MINIMUM CANARY FLOOR  : NCCL AllReduce >= 400.0 GB/s (BusBw)\n")
	doc.WriteString("HARDWARE RELIABILITY  : 0 unrecovered ECC errors; 0 thermal throttling events\n")
	doc.WriteString(fmt.Sprintf("RESERVATION DURATION  : %d Continuous Hours (Take-or-Pay Delivery)\n\n", durationHours))

	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString("3. INSTITUTIONAL INVARIANTS & LEGAL COVENANTS\n")
	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString("[INVARIANT 1 - PHYSICAL FORWARD EXCLUSION]\n")
	doc.WriteString("This Agreement constitutes a bilateral, physically delivered forward reservation of enterprise\n")
	doc.WriteString("compute capacity. It is strictly non-transferable, non-fungible, and does not represent\n")
	doc.WriteString("a continuous order book or cash-settled synthetic perpetual. Delivery is verified via\n")
	doc.WriteString("cryptographic telemetry attestation.\n\n")

	doc.WriteString("[INVARIANT 4 - TELEMETRY ZERO-WORKLOAD PRIVACY]\n")
	doc.WriteString("Verification of capacity is performed strictly through synthetic hardware health canaries\n")
	doc.WriteString("prior to handover. Under no circumstances shall customer workload data, model weights,\n")
	doc.WriteString("training code, or user prompts be ingested or inspected by Verinode telemetry systems.\n\n")

	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString("4. CRYPTOGRAPHIC INTEGRITY DIGEST\n")
	doc.WriteString("--------------------------------------------------------------------------------\n")
	doc.WriteString(fmt.Sprintf("CANONICAL ROOT TRADE ID: %s\n", c.ID))

	contentWithoutChecksum := doc.String()
	hasher := sha256.New()
	hasher.Write([]byte(contentWithoutChecksum))
	checksum := hex.EncodeToString(hasher.Sum(nil))

	finalContent := fmt.Sprintf("%sSHA-256 INTEGRITY DIGEST : %s\n================================================================================\n",
		contentWithoutChecksum, checksum)

	return &ConfirmationDocument{
		TradeID:         c.ID,
		ContractDate:    c.CreatedAt.UTC(),
		Buyer:           buyer,
		Seller:          seller,
		GradeID:         c.GradeID,
		TemplateVersion: c.TemplateVersion,
		DurationHours:   durationHours,
		State:           string(c.State),
		SHA256Checksum:  checksum,
		DocumentContent: finalContent,
	}
}
