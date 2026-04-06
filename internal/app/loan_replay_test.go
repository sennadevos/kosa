package app

import (
	"testing"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

func TestReplayNone(t *testing.T) {
	loan := &domain.Loan{
		OriginalAmount: amountOf("150.00"),
		InterestType:   domain.InterestNone,
	}
	payments := []domain.LoanPayment{
		{Amount: amountOf("75.00"), Date: time.Now()},
	}

	status, _ := ReplayLoan(loan, payments)

	assertAmount(t, "total_owed", "150.00", status.TotalOwed)
	assertAmount(t, "total_paid", "75.00", status.TotalPaid)
	assertAmount(t, "remaining", "75.00", status.Remaining)
	assertAmount(t, "interest", "0.00", status.TotalInterest)
}

func TestReplayNoneFullyPaid(t *testing.T) {
	loan := &domain.Loan{
		OriginalAmount: amountOf("100.00"),
		InterestType:   domain.InterestNone,
	}
	payments := []domain.LoanPayment{
		{Amount: amountOf("60.00"), Date: time.Now()},
		{Amount: amountOf("40.00"), Date: time.Now()},
	}

	status, _ := ReplayLoan(loan, payments)

	assertAmount(t, "remaining", "0.00", status.Remaining)
}

func TestReplayFlat(t *testing.T) {
	loan := &domain.Loan{
		OriginalAmount: amountOf("1000.00"),
		InterestType:   domain.InterestFlat,
		InterestRate:   amountOf("5"),
	}
	payments := []domain.LoanPayment{
		{Amount: amountOf("200.00"), Date: time.Now()},
	}

	status, _ := ReplayLoan(loan, payments)

	assertAmount(t, "interest", "50.00", status.TotalInterest)
	assertAmount(t, "total_owed", "1050.00", status.TotalOwed)
	assertAmount(t, "total_paid", "200.00", status.TotalPaid)
	assertAmount(t, "remaining", "850.00", status.Remaining)
}

func TestReplayPeriodic(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	loan := &domain.Loan{
		OriginalAmount: amountOf("1000.00"),
		InterestType:   domain.InterestPeriodic,
		InterestRate:   amountOf("2"),
		InterestPeriod: domain.PeriodMonthly,
		DateCreated:    start,
	}

	// payment at end of month 1
	payments := []domain.LoanPayment{
		{Amount: amountOf("350.00"), Date: start.AddDate(0, 0, 28)},
	}

	status, timeline := ReplayLoan(loan, payments)

	if len(timeline) == 0 {
		t.Fatal("expected timeline entries")
	}

	// first period: 1000 + 2% = 1020, - 350 = 670
	first := timeline[0]
	assertAmount(t, "opening", "1000.00", first.OpeningBalance)
	assertAmount(t, "interest", "20.00", first.Interest)
	assertAmount(t, "payments", "350.00", first.Payments)
	assertAmount(t, "closing", "670.00", first.ClosingBalance)

	if status.TotalPaid.Format() != "350.00" {
		t.Fatalf("expected paid 350.00, got %s", status.TotalPaid.Format())
	}
}

func TestReplayOverpayment(t *testing.T) {
	loan := &domain.Loan{
		OriginalAmount: amountOf("100.00"),
		InterestType:   domain.InterestNone,
	}
	payments := []domain.LoanPayment{
		{Amount: amountOf("120.00"), Date: time.Now()},
	}

	status, _ := ReplayLoan(loan, payments)
	assertAmount(t, "remaining", "0.00", status.Remaining)
}

func TestSplit(t *testing.T) {
	a, _ := testApp()
	ctx := t.Context()

	total, _ := domain.NewAmount("400.00")
	result, err := a.Split(ctx, SplitInput{
		TotalAmount: total,
		Description: "Airbnb",
		FriendNames: []string{"Bas", "Jan", "Lisa"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.PersonalExpense == nil {
		t.Fatal("expected personal expense")
	}
	if len(result.Loans) != 3 {
		t.Fatalf("expected 3 loans, got %d", len(result.Loans))
	}
	if len(result.LoanExpenses) != 3 {
		t.Fatalf("expected 3 loan expenses, got %d", len(result.LoanExpenses))
	}

	// check each loan expense is loan-linked
	for _, txn := range result.LoanExpenses {
		if txn.LoanID == "" {
			t.Fatal("expected loan-linked expense")
		}
	}
}

func amountOf(s string) domain.Amount {
	a, _ := domain.NewAmount(s)
	return a
}

func assertAmount(t *testing.T, label, expected string, actual domain.Amount) {
	t.Helper()
	if actual.Format() != expected {
		t.Fatalf("%s: expected %s, got %s", label, expected, actual.Format())
	}
}
