package domain

import "time"

type LoanPayment struct {
	ID        string
	LoanID    string
	Date      time.Time
	Amount    Amount
	AccountID string
	AccountName string
	Notes     string
}

type LoanPaymentInput struct {
	LoanID    string
	Date      time.Time
	Amount    Amount
	AccountID string
	Notes     string
}
