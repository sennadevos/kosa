package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type IncomeInput struct {
	Amount          domain.Amount
	Description     string
	Category        string
	Account         string
	Tags            []string
	Date            time.Time
	ForeignAmount   domain.Amount
	ForeignCurrency string
	Reference       string
	Notes           string
}

func (a *App) Income(ctx context.Context, in IncomeInput) (*domain.Transaction, error) {
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

	var exchangeRate domain.Amount
	if !in.ForeignAmount.IsZero() && !in.Amount.IsZero() {
		exchangeRate = in.Amount.Div(in.ForeignAmount)
	}

	t := domain.TransactionInput{
		Date:            date,
		Type:            domain.TransactionIncome,
		Amount:          in.Amount,
		Description:     in.Description,
		CategoryID:      catID,
		TagIDs:          tagIDs,
		AccountID:       accountID,
		ForeignAmount:   in.ForeignAmount,
		ForeignCurrency: in.ForeignCurrency,
		ExchangeRate:    exchangeRate,
		Reference:       in.Reference,
		Notes:           in.Notes,
	}

	txn, err := a.Backend.CreateTransaction(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating income: %w", err)
	}
	return txn, nil
}
