package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type RecurringAddInput struct {
	Name       string
	Type       domain.TransactionType
	Amount     domain.Amount
	Category   string
	Tags       []string
	Account    string
	Frequency  domain.Frequency
	DayOfMonth int
	StartDate  time.Time
	EndDate    *time.Time
	Notes      string
}

func (a *App) RecurringAdd(ctx context.Context, in RecurringAddInput) (*domain.RecurringRule, error) {
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

	startDate := in.StartDate
	if startDate.IsZero() {
		startDate = time.Now()
	}

	rule, err := a.Backend.CreateRecurringRule(ctx, domain.RecurringRuleInput{
		Name:       in.Name,
		Type:       in.Type,
		Amount:     in.Amount,
		CategoryID: catID,
		TagIDs:     tagIDs,
		AccountID:  accountID,
		Frequency:  in.Frequency,
		DayOfMonth: in.DayOfMonth,
		StartDate:  startDate,
		EndDate:    in.EndDate,
		IsActive:   true,
		Notes:      in.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("creating recurring rule: %w", err)
	}
	return rule, nil
}

func (a *App) RecurringList(ctx context.Context) ([]domain.RecurringRule, error) {
	return a.Backend.ListRecurringRules(ctx, domain.RecurringRuleFilter{})
}

func (a *App) RecurringPause(ctx context.Context, id string) error {
	rule, err := a.Backend.GetRecurringRule(ctx, id)
	if err != nil {
		return err
	}
	_, err = a.Backend.UpdateRecurringRule(ctx, id, domain.RecurringRuleInput{
		Name:       rule.Name,
		Type:       rule.Type,
		Amount:     rule.Amount,
		CategoryID: rule.CategoryID,
		TagIDs:     rule.TagIDs,
		AccountID:  rule.AccountID,
		Frequency:  rule.Frequency,
		DayOfMonth: rule.DayOfMonth,
		StartDate:  rule.StartDate,
		EndDate:    rule.EndDate,
		IsActive:   false,
		Notes:      rule.Notes,
	})
	return err
}

func (a *App) RecurringResume(ctx context.Context, id string) error {
	rule, err := a.Backend.GetRecurringRule(ctx, id)
	if err != nil {
		return err
	}
	_, err = a.Backend.UpdateRecurringRule(ctx, id, domain.RecurringRuleInput{
		Name:       rule.Name,
		Type:       rule.Type,
		Amount:     rule.Amount,
		CategoryID: rule.CategoryID,
		TagIDs:     rule.TagIDs,
		AccountID:  rule.AccountID,
		Frequency:  rule.Frequency,
		DayOfMonth: rule.DayOfMonth,
		StartDate:  rule.StartDate,
		EndDate:    rule.EndDate,
		IsActive:   true,
		Notes:      rule.Notes,
	})
	return err
}
