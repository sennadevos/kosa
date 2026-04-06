package app

import (
	"context"
	"strings"

	"github.com/sennadevos/kosa/internal/domain"
)

type ListInput struct {
	Limit    int
	Category string
	Tag      string
	Account  string
	Type     string
	DateFrom string
	DateTo   string
}

func (a *App) ListTransactions(ctx context.Context, in ListInput) ([]domain.Transaction, error) {
	filter := domain.TransactionFilter{
		Limit: in.Limit,
	}

	if in.Category != "" {
		catID, err := a.resolveCategoryID(ctx, in.Category)
		if err != nil {
			return nil, err
		}
		filter.CategoryID = catID
	}

	if in.Account != "" {
		accID, err := a.ResolveAccountID(ctx, in.Account)
		if err != nil {
			return nil, err
		}
		filter.AccountID = accID
	}

	if in.Type != "" {
		t := domain.TransactionType(in.Type)
		filter.Type = &t
	}

	return a.Backend.ListTransactions(ctx, filter)
}

func (a *App) SearchTransactions(ctx context.Context, query string, limit int) ([]domain.Transaction, error) {
	filter := domain.TransactionFilter{
		Limit: limit,
	}
	all, err := a.Backend.ListTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matched []domain.Transaction
	for _, t := range all {
		if strings.Contains(strings.ToLower(t.Description), query) {
			matched = append(matched, t)
		}
	}
	if limit > 0 && len(matched) > limit {
		matched = matched[len(matched)-limit:]
	}
	return matched, nil
}
