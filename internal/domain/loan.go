package domain

import "time"

type LoanType string

const (
	LoanPayable    LoanType = "payable"
	LoanReceivable LoanType = "receivable"
)

type InterestType string

const (
	InterestNone     InterestType = "none"
	InterestFlat     InterestType = "flat"
	InterestPeriodic InterestType = "periodic"
)

type InterestPeriod string

const (
	PeriodWeekly    InterestPeriod = "weekly"
	PeriodMonthly   InterestPeriod = "monthly"
	PeriodQuarterly InterestPeriod = "quarterly"
	PeriodYearly    InterestPeriod = "yearly"
)

type Loan struct {
	ID               string
	Type             LoanType
	CounterpartyName string
	CounterpartyURI  string
	Description      string
	OriginalAmount   Amount
	Currency         string
	DateCreated      time.Time
	DueDate          *time.Time
	InterestType     InterestType
	InterestRate     Amount
	InterestPeriod   InterestPeriod
	IsSettled        bool
	Notes            string
}

type LoanInput struct {
	Type             LoanType
	CounterpartyName string
	CounterpartyURI  string
	Description      string
	OriginalAmount   Amount
	Currency         string
	DateCreated      time.Time
	DueDate          *time.Time
	InterestType     InterestType
	InterestRate     Amount
	InterestPeriod   InterestPeriod
	IsSettled        bool
	Notes            string
}

type LoanFilter struct {
	Type      *LoanType
	Settled   *bool
	CounterpartyName string
}

// LoanStatus holds computed fields derived at query time.
type LoanStatus struct {
	TotalInterest Amount
	TotalOwed     Amount
	TotalPaid     Amount
	Remaining     Amount
	IsOverdue     bool
}

// LoanTimelineEntry represents one period in a loan replay.
type LoanTimelineEntry struct {
	PeriodEnd      time.Time
	OpeningBalance Amount
	Interest       Amount
	Payments       Amount
	ClosingBalance Amount
}
