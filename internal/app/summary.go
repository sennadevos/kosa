package app

import (
	"context"
	"sort"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type CategorySummary struct {
	CategoryID   string
	CategoryName string
	Expenses     domain.Amount
	Refunds      domain.Amount
	Net          domain.Amount
}

// SpendingSummary computes net spending per category for a date range.
// Excludes loan-linked transactions.
func (a *App) SpendingSummary(ctx context.Context, from, to time.Time) ([]CategorySummary, error) {
	txns, err := a.Backend.ListTransactions(ctx, domain.TransactionFilter{
		DateFrom: &from,
		DateTo:   &to,
	})
	if err != nil {
		return nil, err
	}

	// load categories for name resolution
	cats, _ := a.Backend.ListCategories(ctx)
	catNames := make(map[string]string, len(cats))
	for _, c := range cats {
		catNames[c.ID] = c.Name
	}

	// accumulate per category
	type accumulator struct {
		expenses domain.Amount
		refunds  domain.Amount
	}
	byCategory := make(map[string]*accumulator)

	for _, t := range txns {
		// skip loan-linked transactions
		if t.LoanID != "" {
			continue
		}
		// skip transfers
		if t.Type == domain.TransactionTransfer {
			continue
		}

		catID := t.CategoryID
		if catID == "" {
			catID = "_uncategorized"
		}

		acc, ok := byCategory[catID]
		if !ok {
			acc = &accumulator{
				expenses: domain.ZeroAmount(),
				refunds:  domain.ZeroAmount(),
			}
			byCategory[catID] = acc
		}

		switch t.Type {
		case domain.TransactionExpense:
			acc.expenses = acc.expenses.Add(t.Amount)
		case domain.TransactionRefund:
			acc.refunds = acc.refunds.Add(t.Amount)
		}
	}

	var summaries []CategorySummary
	for catID, acc := range byCategory {
		name := catNames[catID]
		if name == "" && catID == "_uncategorized" {
			name = "uncategorized"
		}
		net := acc.expenses.Sub(acc.refunds)
		summaries = append(summaries, CategorySummary{
			CategoryID:   catID,
			CategoryName: name,
			Expenses:     acc.expenses,
			Refunds:      acc.refunds,
			Net:          net,
		})
	}

	// sort by net descending
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Net.GreaterThan(summaries[j].Net.Decimal)
	})

	return summaries, nil
}
