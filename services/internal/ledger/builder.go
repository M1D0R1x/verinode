package ledger

import (
	"errors"
	"time"
)

// NewBuyerDepositTransaction creates a balanced transaction for an incoming buyer deposit into escrow.
// Entry 1: Debit Escrow Cash Asset (+amount)
// Entry 2: Credit Buyer Deposit Liability (-amount)
// Net Sum == 0
func NewBuyerDepositTransaction(
	txID string,
	escrowAccountID string,
	buyerDepositAccountID string,
	contractID string,
	amountCents int64,
	currency string,
	bankRef string,
) ([]Entry, error) {
	if amountCents <= 0 {
		return nil, errors.New("ledger: deposit amount must be positive")
	}

	now := time.Now().UTC()

	entries := []Entry{
		{
			TransactionID: txID,
			AccountID:     escrowAccountID,
			ContractID:    contractID,
			AmountCents:   amountCents, // Debit (+): Cash received into platform escrow
			Currency:      currency,
			BankReference: bankRef,
			CreatedAt:     now,
		},
		{
			TransactionID: txID,
			AccountID:     buyerDepositAccountID,
			ContractID:    contractID,
			AmountCents:   -amountCents, // Credit (-): Liability owed to buyer
			Currency:      currency,
			BankReference: bankRef,
			CreatedAt:     now,
		},
	}

	if err := ValidateTransaction(entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// NewSettlementTransaction creates a balanced 3-legged settlement transaction:
// Entry 1: Credit Escrow Cash (-totalAmount)
// Entry 2: Debit Seller Payout (+sellerAmount)
// Entry 3: Debit Platform Fee Revenue (+platformFee)
// Net Sum == 0 (-totalAmount + sellerAmount + platformFee == 0)
func NewSettlementTransaction(
	txID string,
	escrowAccountID string,
	sellerPayoutAccountID string,
	platformFeeAccountID string,
	contractID string,
	totalAmountCents int64,
	feeCents int64,
	currency string,
) ([]Entry, error) {
	if totalAmountCents <= 0 {
		return nil, errors.New("ledger: settlement total must be positive")
	}
	if feeCents < 0 || feeCents >= totalAmountCents {
		return nil, errors.New("ledger: invalid fee amount relative to settlement total")
	}

	sellerAmount := totalAmountCents - feeCents
	now := time.Now().UTC()

	entries := []Entry{
		{
			TransactionID: txID,
			AccountID:     escrowAccountID,
			ContractID:    contractID,
			AmountCents:   -totalAmountCents, // Credit (-): Outflow from escrow
			Currency:      currency,
			CreatedAt:     now,
		},
		{
			TransactionID: txID,
			AccountID:     sellerPayoutAccountID,
			ContractID:    contractID,
			AmountCents:   sellerAmount, // Debit (+): Net payout obligation to seller
			Currency:      currency,
			CreatedAt:     now,
		},
		{
			TransactionID: txID,
			AccountID:     platformFeeAccountID,
			ContractID:    contractID,
			AmountCents:   feeCents, // Debit (+): Marketplace brokerage revenue
			Currency:      currency,
			CreatedAt:     now,
		},
	}

	// If fee is 0, filter out the zero-entry to maintain non-zero entry rule
	if feeCents == 0 {
		entries = entries[:2]
	}

	if err := ValidateTransaction(entries); err != nil {
		return nil, err
	}

	return entries, nil
}
