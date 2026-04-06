package app

import (
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

// GenerateProjections produces synthetic transactions from recurring rules
// for a given date range. These are virtual — never stored.
func GenerateProjections(rules []domain.RecurringRule, from, to time.Time) []domain.Transaction {
	var projections []domain.Transaction

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		dates := occurrences(rule, from, to)
		for _, d := range dates {
			projections = append(projections, domain.Transaction{
				Date:        d,
				Type:        rule.Type,
				Amount:      rule.Amount,
				Description: rule.Name,
				CategoryID:  rule.CategoryID,
				AccountID:   rule.AccountID,
			})
		}
	}

	return projections
}

// occurrences returns all dates a rule fires between from and to (inclusive).
func occurrences(rule domain.RecurringRule, from, to time.Time) []time.Time {
	var dates []time.Time

	// start from rule's start date or from, whichever is later
	cursor := rule.StartDate
	if cursor.Before(from) {
		cursor = alignToNext(cursor, rule, from)
	}

	for !cursor.After(to) {
		if rule.EndDate != nil && cursor.After(*rule.EndDate) {
			break
		}
		if !cursor.Before(from) {
			dates = append(dates, cursor)
		}
		cursor = advance(cursor, rule)
	}

	return dates
}

func advance(t time.Time, rule domain.RecurringRule) time.Time {
	switch rule.Frequency {
	case domain.FreqDaily:
		return t.AddDate(0, 0, 1)
	case domain.FreqWeekly:
		return t.AddDate(0, 0, 7)
	case domain.FreqBiweekly:
		return t.AddDate(0, 0, 14)
	case domain.FreqMonthly:
		next := t.AddDate(0, 1, 0)
		if rule.DayOfMonth > 0 {
			next = time.Date(next.Year(), next.Month(), clampDay(rule.DayOfMonth, next), 0, 0, 0, 0, t.Location())
		}
		return next
	case domain.FreqQuarterly:
		return t.AddDate(0, 3, 0)
	case domain.FreqYearly:
		return t.AddDate(1, 0, 0)
	default:
		return t.AddDate(0, 1, 0)
	}
}

func alignToNext(start time.Time, rule domain.RecurringRule, target time.Time) time.Time {
	cursor := start
	for cursor.Before(target) {
		cursor = advance(cursor, rule)
	}
	return cursor
}

func clampDay(day int, t time.Time) int {
	maxDay := daysInMonth(t.Year(), t.Month())
	if day > maxDay {
		return maxDay
	}
	return day
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
