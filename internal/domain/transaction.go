package domain

import "time"

type TransactionType string

const (
	TransactionIncome   TransactionType = "income"
	TransactionExpense  TransactionType = "expense"
	TransactionTransfer TransactionType = "transfer"
	TransactionRefund   TransactionType = "refund"
)

type Transaction struct {
	ID              string
	Date            time.Time
	Type            TransactionType
	Amount          Amount
	Description     string
	CategoryID      string
	CategoryName    string
	TagIDs          []string
	TagNames        []string
	AccountID       string
	AccountName     string
	ToAccountID     string
	ToAccountName   string
	LoanID          string
	RecurringRuleID string
	RefundOfID      string
	Cashback        Amount
	Reference       string
	ForeignAmount   Amount
	ForeignCurrency string
	ExchangeRate    Amount
	Notes           string
}

type TransactionInput struct {
	Date            time.Time
	Type            TransactionType
	Amount          Amount
	Description     string
	CategoryID      string
	TagIDs          []string
	AccountID       string
	ToAccountID     string
	LoanID          string
	RecurringRuleID string
	RefundOfID      string
	Cashback        Amount
	Reference       string
	ForeignAmount   Amount
	ForeignCurrency string
	ExchangeRate    Amount
	Notes           string
}

type TransactionFilter struct {
	AccountID  string
	CategoryID string
	TagID      string
	Type       *TransactionType
	LoanID     string
	DateFrom   *time.Time
	DateTo     *time.Time
	Limit      int
}
