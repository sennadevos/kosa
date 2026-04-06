package domain

import "time"

type SnapshotSource string

const (
	SourceManual         SnapshotSource = "manual"
	SourceBankImport     SnapshotSource = "bank_import"
	SourceReconciliation SnapshotSource = "reconciliation"
)

type BalanceSnapshot struct {
	ID        string
	AccountID string
	AccountName string
	Date      time.Time
	Balance   Amount
	Source    SnapshotSource
	Notes     string
}

type SnapshotInput struct {
	AccountID string
	Date      time.Time
	Balance   Amount
	Source    SnapshotSource
	Notes     string
}
