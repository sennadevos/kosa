package domain

import "time"

type Frequency string

const (
	FreqDaily     Frequency = "daily"
	FreqWeekly    Frequency = "weekly"
	FreqBiweekly  Frequency = "biweekly"
	FreqMonthly   Frequency = "monthly"
	FreqQuarterly Frequency = "quarterly"
	FreqYearly    Frequency = "yearly"
)

type RecurringRule struct {
	ID          string
	Name        string
	Type        TransactionType
	Amount      Amount
	CategoryID  string
	CategoryName string
	TagIDs      []string
	TagNames    []string
	AccountID   string
	AccountName string
	Frequency   Frequency
	DayOfMonth  int
	StartDate   time.Time
	EndDate     *time.Time
	IsActive    bool
	Notes       string
}

type RecurringRuleInput struct {
	Name       string
	Type       TransactionType
	Amount     Amount
	CategoryID string
	TagIDs     []string
	AccountID  string
	Frequency  Frequency
	DayOfMonth int
	StartDate  time.Time
	EndDate    *time.Time
	IsActive   bool
	Notes      string
}

type RecurringRuleFilter struct {
	ActiveOnly bool
}
