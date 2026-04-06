package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type TransferInput struct {
	Amount      domain.Amount
	FromAccount string
	ToAccount   string
	Date        time.Time
	Notes       string
}

func (a *App) Transfer(ctx context.Context, in TransferInput) (*domain.Transaction, error) {
	fromID, err := a.ResolveAccountID(ctx, in.FromAccount)
	if err != nil {
		return nil, fmt.Errorf("from account: %w", err)
	}
	toID, err := a.ResolveAccountID(ctx, in.ToAccount)
	if err != nil {
		return nil, fmt.Errorf("to account: %w", err)
	}

	date := in.Date
	if date.IsZero() {
		date = time.Now()
	}

	t := domain.TransactionInput{
		Date:        date,
		Type:        domain.TransactionTransfer,
		Amount:      in.Amount,
		Description: fmt.Sprintf("Transfer"),
		AccountID:   fromID,
		ToAccountID: toID,
		Notes:       in.Notes,
	}

	txn, err := a.Backend.CreateTransaction(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating transfer: %w", err)
	}
	return txn, nil
}
