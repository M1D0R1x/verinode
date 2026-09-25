package ledger

import (
	"fmt"
	"time"
)

// AccountType classifies the role of a subledger account.
type AccountType string

const (
	AccountTypeDeposit    AccountType = "deposit"
	AccountTypePayable    AccountType = "payable"
	AccountTypeReceivable AccountType = "receivable"
	AccountTypeCollateral AccountType = "collateral"
)

var validAccountTypes = map[AccountType]bool{
	AccountTypeDeposit:    true,
	AccountTypePayable:    true,
	AccountTypeReceivable: true,
	AccountTypeCollateral: true,
}

func (at AccountType) IsValid() bool {
	return validAccountTypes[at]
}

// Account represents a financial balance container tied to a participant.
type Account struct {
	ID            string      `json:"id"`
	ParticipantID string      `json:"participant_id"`
	Currency      string      `json:"currency"`
	AccountType   AccountType `json:"account_type"`
	CreatedAt     time.Time   `json:"created_at"`
}

func ParseAccountType(s string) (AccountType, error) {
	at := AccountType(s)
	if !at.IsValid() {
		return "", fmt.Errorf("invalid account type: %q", s)
	}
	return at, nil
}
