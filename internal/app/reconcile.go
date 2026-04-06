package app

import (
	"context"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type ReconcileRow struct {
	RuleName string
	Expected domain.Amount
	Actual   domain.Amount
	Delta    domain.Amount
	Status   string // "linked", "missing", "unlinked"
}

// Reconcile compares actual transactions to recurring rule projections for a given month.
func (a *App) Reconcile(ctx context.Context, year int, month time.Month) ([]ReconcileRow, error) {
	// get all active rules
	rules, err := a.Backend.ListRecurringRules(ctx, domain.RecurringRuleFilter{ActiveOnly: true})
	if err != nil {
		return nil, err
	}

	// date range for the month
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)

	// get actual transactions for the month
	txns, err := a.Backend.ListTransactions(ctx, domain.TransactionFilter{
		DateFrom: &from,
		DateTo:   &to,
	})
	if err != nil {
		return nil, err
	}

	// build map of rule ID -> actual transaction
	ruleToTxn := make(map[string]*domain.Transaction)
	for i, t := range txns {
		if t.RecurringRuleID != "" {
			ruleToTxn[t.RecurringRuleID] = &txns[i]
		}
	}

	var rows []ReconcileRow
	for _, rule := range rules {
		// check if this rule has an occurrence in this month
		projected := GenerateProjections([]domain.RecurringRule{rule}, from, to)
		if len(projected) == 0 {
			continue
		}

		expected := rule.Amount
		txn, linked := ruleToTxn[rule.ID]

		if linked {
			delta := txn.Amount.Sub(expected)
			rows = append(rows, ReconcileRow{
				RuleName: rule.Name,
				Expected: expected,
				Actual:   txn.Amount,
				Delta:    delta,
				Status:   "linked",
			})
		} else {
			rows = append(rows, ReconcileRow{
				RuleName: rule.Name,
				Expected: expected,
				Actual:   domain.ZeroAmount(),
				Delta:    domain.ZeroAmount(),
				Status:   "missing",
			})
		}
	}

	return rows, nil
}
