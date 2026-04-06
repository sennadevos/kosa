package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type AccountBalance struct {
	AccountID   string
	AccountName string
	Balance     domain.Amount
}

// CurrentBalance computes the balance for an account: latest snapshot + transactions since.
func (a *App) CurrentBalance(ctx context.Context, accountID string) (domain.Amount, error) {
	snap, err := a.Backend.LatestSnapshot(ctx, accountID)
	if err != nil {
		return domain.ZeroAmount(), fmt.Errorf("no snapshot found for account: %w", err)
	}

	// get transactions since snapshot
	txns, err := a.Backend.ListTransactions(ctx, domain.TransactionFilter{
		AccountID: accountID,
		DateFrom:  &snap.Date,
	})
	if err != nil {
		return domain.ZeroAmount(), err
	}

	balance := snap.Balance
	for _, t := range txns {
		balance = applyTransaction(balance, t, accountID)
	}

	return balance, nil
}

// ProjectedBalance computes what the balance will be on a future date.
func (a *App) ProjectedBalance(ctx context.Context, accountID string, targetDate time.Time) (domain.Amount, error) {
	current, err := a.CurrentBalance(ctx, accountID)
	if err != nil {
		return domain.ZeroAmount(), err
	}

	rules, err := a.Backend.ListRecurringRules(ctx, domain.RecurringRuleFilter{ActiveOnly: true})
	if err != nil {
		return domain.ZeroAmount(), err
	}

	// filter rules for this account
	var accountRules []domain.RecurringRule
	for _, r := range rules {
		if r.AccountID == accountID || r.AccountID == "" {
			accountRules = append(accountRules, r)
		}
	}

	projections := GenerateProjections(accountRules, time.Now(), targetDate)
	for _, p := range projections {
		current = applyTransaction(current, p, accountID)
	}

	return current, nil
}

// AllBalances returns current balance for every account.
func (a *App) AllBalances(ctx context.Context) ([]AccountBalance, error) {
	accounts, err := a.Backend.ListAccounts(ctx, domain.AccountFilter{})
	if err != nil {
		return nil, err
	}

	var balances []AccountBalance
	for _, acc := range accounts {
		bal, err := a.CurrentBalance(ctx, acc.ID)
		if err != nil {
			// no snapshot yet — skip
			bal = domain.ZeroAmount()
		}
		balances = append(balances, AccountBalance{
			AccountID:   acc.ID,
			AccountName: acc.Name,
			Balance:     bal,
		})
	}

	return balances, nil
}

func applyTransaction(balance domain.Amount, t domain.Transaction, accountID string) domain.Amount {
	switch t.Type {
	case domain.TransactionIncome, domain.TransactionRefund:
		return balance.Add(t.Amount)
	case domain.TransactionExpense:
		return balance.Sub(t.Amount)
	case domain.TransactionTransfer:
		if t.AccountID == accountID {
			return balance.Sub(t.Amount) // money leaving
		}
		if t.ToAccountID == accountID {
			return balance.Add(t.Amount) // money arriving
		}
	}
	return balance
}
