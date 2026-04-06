package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type RefundInput struct {
	Amount      domain.Amount
	Description string
	Category    string
	Account     string
	Tags        []string
	Date        time.Time
	RefundOfID  string
	Notes       string
}

func (a *App) Refund(ctx context.Context, in RefundInput) (*domain.Transaction, error) {
	// validate original transaction exists if specified
	if in.RefundOfID != "" {
		_, err := a.Backend.GetTransaction(ctx, in.RefundOfID)
		if err != nil {
			return nil, fmt.Errorf("original transaction: %w", err)
		}
	}

	accountID, err := a.ResolveAccountID(ctx, in.Account)
	if err != nil {
		return nil, err
	}
	catID, err := a.resolveCategoryID(ctx, in.Category)
	if err != nil {
		return nil, err
	}
	tagIDs, err := a.resolveTagIDs(ctx, in.Tags)
	if err != nil {
		return nil, err
	}

	date := in.Date
	if date.IsZero() {
		date = time.Now()
	}

	t := domain.TransactionInput{
		Date:        date,
		Type:        domain.TransactionRefund,
		Amount:      in.Amount,
		Description: in.Description,
		CategoryID:  catID,
		TagIDs:      tagIDs,
		AccountID:   accountID,
		RefundOfID:  in.RefundOfID,
		Notes:       in.Notes,
	}

	txn, err := a.Backend.CreateTransaction(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating refund: %w", err)
	}
	return txn, nil
}
