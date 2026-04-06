package app

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sennadevos/kosa/internal/domain"
)

// ReplayLoan computes the current status and timeline of a loan from its
// terms and payment history. This is a pure function — no database access.
func ReplayLoan(loan *domain.Loan, payments []domain.LoanPayment) (*domain.LoanStatus, []domain.LoanTimelineEntry) {
	switch loan.InterestType {
	case domain.InterestPeriodic:
		return replayPeriodic(loan, payments)
	case domain.InterestFlat:
		return replayFlat(loan, payments)
	default:
		return replayNone(loan, payments)
	}
}

func replayNone(loan *domain.Loan, payments []domain.LoanPayment) (*domain.LoanStatus, []domain.LoanTimelineEntry) {
	totalPaid := domain.ZeroAmount()
	for _, p := range payments {
		totalPaid = totalPaid.Add(p.Amount)
	}

	totalOwed := loan.OriginalAmount
	remaining := totalOwed.Sub(totalPaid)
	if remaining.IsNegative() {
		remaining = domain.ZeroAmount()
	}

	status := &domain.LoanStatus{
		TotalInterest: domain.ZeroAmount(),
		TotalOwed:     totalOwed,
		TotalPaid:     totalPaid,
		Remaining:     remaining,
		IsOverdue:     loan.DueDate != nil && loan.DueDate.Before(time.Now()) && !loan.IsSettled,
	}

	return status, nil
}

func replayFlat(loan *domain.Loan, payments []domain.LoanPayment) (*domain.LoanStatus, []domain.LoanTimelineEntry) {
	totalPaid := domain.ZeroAmount()
	for _, p := range payments {
		totalPaid = totalPaid.Add(p.Amount)
	}

	rate := loan.InterestRate.Div(domain.NewAmountFromInt(100))
	interest := loan.OriginalAmount.Mul(rate)
	totalOwed := loan.OriginalAmount.Add(interest)
	remaining := totalOwed.Sub(totalPaid)
	if remaining.IsNegative() {
		remaining = domain.ZeroAmount()
	}

	status := &domain.LoanStatus{
		TotalInterest: interest,
		TotalOwed:     totalOwed,
		TotalPaid:     totalPaid,
		Remaining:     remaining,
		IsOverdue:     loan.DueDate != nil && loan.DueDate.Before(time.Now()) && !loan.IsSettled,
	}

	return status, nil
}

func replayPeriodic(loan *domain.Loan, payments []domain.LoanPayment) (*domain.LoanStatus, []domain.LoanTimelineEntry) {
	// sort payments by date
	sorted := make([]domain.LoanPayment, len(payments))
	copy(sorted, payments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	rate := loan.InterestRate.Decimal.Div(decimal.NewFromInt(100))
	balance := loan.OriginalAmount.Decimal
	totalInterest := decimal.Zero
	totalPaid := decimal.Zero
	payIdx := 0

	var timeline []domain.LoanTimelineEntry

	// iterate period by period from start to today
	now := time.Now()
	periodStart := loan.DateCreated
	for {
		periodEnd := nextPeriodBoundary(periodStart, loan.InterestPeriod)
		if periodEnd.After(now) {
			periodEnd = now
		}

		opening := balance

		// accrue interest at period boundary
		interest := balance.Mul(rate)
		balance = balance.Add(interest)
		totalInterest = totalInterest.Add(interest)

		// apply payments within this period
		periodPayments := decimal.Zero
		for payIdx < len(sorted) && !sorted[payIdx].Date.After(periodEnd) {
			balance = balance.Sub(sorted[payIdx].Amount.Decimal)
			totalPaid = totalPaid.Add(sorted[payIdx].Amount.Decimal)
			periodPayments = periodPayments.Add(sorted[payIdx].Amount.Decimal)
			payIdx++
		}

		if balance.IsNegative() {
			balance = decimal.Zero
		}

		timeline = append(timeline, domain.LoanTimelineEntry{
			PeriodEnd:      periodEnd,
			OpeningBalance: domain.Amount{Decimal: opening},
			Interest:       domain.Amount{Decimal: interest},
			Payments:       domain.Amount{Decimal: periodPayments},
			ClosingBalance: domain.Amount{Decimal: balance},
		})

		if periodEnd.Equal(now) || periodEnd.After(now) || balance.IsZero() {
			break
		}
		periodStart = periodEnd
	}

	remaining := domain.Amount{Decimal: balance}
	if remaining.IsNegative() {
		remaining = domain.ZeroAmount()
	}

	status := &domain.LoanStatus{
		TotalInterest: domain.Amount{Decimal: totalInterest},
		TotalOwed:     loan.OriginalAmount.Add(domain.Amount{Decimal: totalInterest}),
		TotalPaid:     domain.Amount{Decimal: totalPaid},
		Remaining:     remaining,
		IsOverdue:     loan.DueDate != nil && loan.DueDate.Before(now) && !loan.IsSettled,
	}

	return status, timeline
}

func nextPeriodBoundary(from time.Time, period domain.InterestPeriod) time.Time {
	switch period {
	case domain.PeriodWeekly:
		return from.AddDate(0, 0, 7)
	case domain.PeriodMonthly:
		return from.AddDate(0, 1, 0)
	case domain.PeriodQuarterly:
		return from.AddDate(0, 3, 0)
	case domain.PeriodYearly:
		return from.AddDate(1, 0, 0)
	default:
		return from.AddDate(0, 1, 0) // default monthly
	}
}
