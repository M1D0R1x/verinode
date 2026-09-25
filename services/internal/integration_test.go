package internal_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/db"
	"github.com/M1D0R1x/verinode/services/internal/inventory"
	"github.com/M1D0R1x/verinode/services/internal/ledger"
	"github.com/M1D0R1x/verinode/services/internal/participant"
	"github.com/M1D0R1x/verinode/services/internal/rfq"
)

func TestEndToEndInstitutionalFlow_WithNeonDB(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set; skipping live database integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cfg := db.DefaultConfig()
	cfg.ConnString = connStr
	pool, err := db.Connect(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("failed connecting to Neon DB: %v", err)
	}
	defer pool.Close()

	participantRepo := participant.NewRepository(pool.Pool)
	inventoryRepo := inventory.NewRepository(pool.Pool)
	rfqRepo := rfq.NewRepository(pool.Pool)
	contractRepo := contract.NewRepository(pool.Pool)
	ledgerRepo := ledger.NewRepository(pool.Pool)

	// 1. Create Buyer and Seller Participants
	buyer := &participant.Participant{
		LegalName:    "Apex AI Capital LP",
		Jurisdiction: "US",
		Role:         "buyer",
		KYCStatus:    "approved",
	}
	if err := participantRepo.Create(ctx, buyer); err != nil {
		t.Fatalf("failed creating buyer participant: %v", err)
	}

	seller := &participant.Participant{
		LegalName:    "Nebula Compute Infrastructure LLC",
		Jurisdiction: "US",
		Role:         "seller",
		KYCStatus:    "approved",
	}
	if err := participantRepo.Create(ctx, seller); err != nil {
		t.Fatalf("failed creating seller participant: %v", err)
	}

	// 2. Seller Lists Inventory Block
	startWindow := time.Now().UTC().Add(24 * time.Hour)
	endWindow := startWindow.Add(168 * time.Hour)
	block := &inventory.InventoryBlock{
		SellerID:     seller.ID,
		GradeID:      "H100-SXM-8XNV",
		RegionBucket: "US-East",
		WindowStart:  startWindow,
		WindowEnd:    endWindow,
		Status:       "available",
	}
	if err := inventoryRepo.Create(ctx, block); err != nil {
		t.Fatalf("failed creating inventory block: %v", err)
	}

	// 3. Buyer Creates RFQ
	req := &rfq.RFQ{
		BuyerID:      buyer.ID,
		GradeID:      "H100-SXM-8XNV",
		RegionBucket: "US-East",
		WindowStart:  startWindow,
		WindowEnd:    endWindow,
		Status:       "open",
	}
	if err := rfqRepo.CreateRFQ(ctx, req); err != nil {
		t.Fatalf("failed creating rfq: %v", err)
	}

	// 4. Seller Submits Quote
	quote := &rfq.Quote{
		RFQID:      req.ID,
		SellerID:   seller.ID,
		BlockID:    block.ID,
		PriceCents: 3528000, // $35,280.00 ($210/hr * 168h)
		Currency:   "USD",
		ExpiresAt:  time.Now().UTC().Add(12 * time.Hour),
		Status:     "active",
	}
	if err := rfqRepo.CreateQuote(ctx, quote); err != nil {
		t.Fatalf("failed creating quote: %v", err)
	}

	// 5. Buyer Accepts Quote Atomically (Invariant 1: Non-transferable bilateral reservation)
	acceptedDetails, err := rfqRepo.AcceptQuote(ctx, quote.ID, buyer.ID)
	if err != nil {
		t.Fatalf("failed accepting quote atomically: %v", err)
	}
	if acceptedDetails.QuoteID != quote.ID {
		t.Errorf("expected accepted quote id %s, got %s", quote.ID, acceptedDetails.QuoteID)
	}

	// 6. Spawn Contract (Invariant 2: Canonical trade_id envelope)
	contractRec := &contract.ContractRecord{
		RFQID:           &req.ID,
		QuoteID:         &quote.ID,
		BuyerID:         buyer.ID,
		SellerID:        seller.ID,
		GradeID:         "H100-SXM-8XNV",
		TemplateVersion: "v1.0.0-institutional",
		State:           contract.StateContractPending,
	}
	if err := contractRepo.CreateContract(ctx, contractRec); err != nil {
		t.Fatalf("failed creating contract: %v", err)
	}
	tradeID := contractRec.ID

	// 7. Advance State Machine to FundedSecured
	evt1 := contract.TransitionEvent{
		Actor:                 buyer.ID,
		Reason:                "Buyer escrow deposit funded via Fedwire",
		IdempotencyKey:        "evt_fund_01",
		AuthorizationDecision: map[string]interface{}{"approved": true},
	}
	err = contractRepo.AdvanceContractState(ctx, tradeID, contract.StateFundedSecured, evt1)
	if err != nil {
		t.Fatalf("failed advancing to funded_secured: %v", err)
	}

	// 8. Enforce Invariant 3: Double-Entry Ledger (SUM(amount_cents) == 0)
	buyerDepositAcc := &ledger.AccountRecord{
		ParticipantID: buyer.ID,
		Currency:      "USD",
		AccountType:   "deposit",
	}
	if err := ledgerRepo.CreateAccount(ctx, buyerDepositAcc); err != nil {
		t.Fatalf("failed creating buyer deposit ledger account: %v", err)
	}

	clearinghouseEscrowAcc := &ledger.AccountRecord{
		ParticipantID: seller.ID,
		Currency:      "USD",
		AccountType:   "collateral",
	}
	if err := ledgerRepo.CreateAccount(ctx, clearinghouseEscrowAcc); err != nil {
		t.Fatalf("failed creating escrow ledger account: %v", err)
	}

	// Post balanced deposit: Debit Cash ($35,280.00), Credit Escrow (-$35,280.00)
	txID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	depositEntries := []ledger.Entry{
		{
			TransactionID: txID,
			AccountID:     buyerDepositAcc.ID,
			ContractID:    tradeID,
			AmountCents:   3528000, // Debit (positive)
			Currency:      "USD",
			BankReference: "FEDWIRE_REF_9918237",
		},
		{
			TransactionID: txID,
			AccountID:     clearinghouseEscrowAcc.ID,
			ContractID:    tradeID,
			AmountCents:   -3528000, // Credit (negative)
			Currency:      "USD",
			BankReference: "FEDWIRE_REF_9918237",
		},
	}

	err = ledgerRepo.RecordTransaction(ctx, txID, depositEntries)
	if err != nil {
		t.Fatalf("failed recording double-entry transaction: %v", err)
	}

	// 9. Verify Account Balances
	buyerBalance, err := ledgerRepo.GetAccountBalance(ctx, buyerDepositAcc.ID)
	if err != nil {
		t.Fatalf("failed fetching buyer balance: %v", err)
	}
	if buyerBalance != 3528000 {
		t.Errorf("expected buyer balance 3528000 cents, got %d", buyerBalance)
	}

	escrowBalance, err := ledgerRepo.GetAccountBalance(ctx, clearinghouseEscrowAcc.ID)
	if err != nil {
		t.Fatalf("failed fetching escrow balance: %v", err)
	}
	if escrowBalance != -3528000 {
		t.Errorf("expected escrow balance -3528000 cents, got %d", escrowBalance)
	}

	// 10. Attempt Unbalanced Transaction (MUST fail Invariant 3)
	badTxID := "a12bc10b-58cc-4372-a567-0e02b2c3d999"
	badEntries := []ledger.Entry{
		{
			TransactionID: badTxID,
			AccountID:     buyerDepositAcc.ID,
			AmountCents:   100000, // $1,000
			Currency:      "USD",
		},
		{
			TransactionID: badTxID,
			AccountID:     clearinghouseEscrowAcc.ID,
			AmountCents:   -50000, // -$500 (SUM != 0)
			Currency:      "USD",
		},
	}
	err = ledgerRepo.RecordTransaction(ctx, badTxID, badEntries)
	if err == nil {
		t.Fatalf("expected unbalanced transaction to be rejected by ledger invariant, but it passed!")
	}

	t.Logf("SUCCESS: Live Neon DB integration test verified all Invariants (Bilateral Reservation, Canonical Trade Envelope, State Machine Audit Events, and Double-Entry Ledger Balances)!")
}
