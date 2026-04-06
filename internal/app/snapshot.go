package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type SnapshotInput struct {
	Balance domain.Amount
	Account string
	Source  domain.SnapshotSource
	Notes   string
}

func (a *App) RecordSnapshot(ctx context.Context, in SnapshotInput) (*domain.BalanceSnapshot, error) {
	accountID, err := a.ResolveAccountID(ctx, in.Account)
	if err != nil {
		return nil, err
	}

	source := in.Source
	if source == "" {
		source = domain.SourceManual
	}

	snap, err := a.Backend.CreateSnapshot(ctx, domain.SnapshotInput{
		AccountID: accountID,
		Date:      time.Now(),
		Balance:   in.Balance,
		Source:    source,
		Notes:     in.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("creating snapshot: %w", err)
	}
	return snap, nil
}
