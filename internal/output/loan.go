package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/sennadevos/kosa/internal/domain"
)

func PrintLoan(w io.Writer, l *domain.Loan, f Format) {
	PrintLoans(w, []domain.Loan{*l}, f)
}

func PrintLoans(w io.Writer, loans []domain.Loan, f Format) {
	switch f {
	case FormatJSON:
		writeJSON(w, loans)
	case FormatToon:
		for _, l := range loans {
			fmt.Fprintf(w, "%s|%s|%s|%s|%s\n",
				l.ID, l.Type, l.OriginalAmount.Format(), l.CounterpartyName, l.Description)
		}
	default:
		printLoansTable(w, loans)
	}
}

func printLoansTable(w io.Writer, loans []domain.Loan) {
	if len(loans) == 0 {
		fmt.Fprintln(w, "no loans")
		return
	}

	fmt.Fprintf(w, "%s  %s  %s  %s  %s  %s\n",
		padRight("id", 10),
		padRight("type", 10),
		padLeft("amount", 10),
		padRight("counterparty", 20),
		padRight("description", 25),
		padRight("settled", 7),
	)
	fmt.Fprintln(w, strings.Repeat("─", 90))

	for _, l := range loans {
		settled := dim + "no" + reset
		if l.IsSettled {
			settled = green + "yes" + reset
		}
		fmt.Fprintf(w, "%s  %s  %s  %s  %s  %s\n",
			padRight(l.ID, 10),
			padRight(string(l.Type), 10),
			padLeft(l.OriginalAmount.Format(), 10),
			padRight(l.CounterpartyName, 20),
			padRight(l.Description, 25),
			settled,
		)
	}
}

func PrintLoanStatus(w io.Writer, l *domain.Loan, s *domain.LoanStatus, f Format) {
	if f == FormatJSON {
		writeJSON(w, map[string]interface{}{
			"loan":   l,
			"status": s,
		})
		return
	}

	fmt.Fprintf(w, "%s — %s (%s)\n", l.Description, l.CounterpartyName, l.Type)
	fmt.Fprintf(w, "  principal:     %s\n", l.OriginalAmount.Format())
	if !s.TotalInterest.IsZero() {
		fmt.Fprintf(w, "  interest:      %s (%s %s at %s%%)\n",
			s.TotalInterest.Format(), l.InterestType, l.InterestPeriod, l.InterestRate.Format())
	}
	fmt.Fprintf(w, "  total owed:    %s\n", s.TotalOwed.Format())
	fmt.Fprintf(w, "  total paid:    %s\n", s.TotalPaid.Format())
	fmt.Fprintf(w, "  remaining:     %s\n", s.Remaining.Format())
	if s.IsOverdue {
		fmt.Fprintf(w, "  %soverdue%s\n", red, reset)
	}
}

func PrintLoanTimeline(w io.Writer, timeline []domain.LoanTimelineEntry, f Format) {
	if f == FormatJSON {
		writeJSON(w, timeline)
		return
	}
	if len(timeline) == 0 {
		return
	}

	fmt.Fprintf(w, "\n%s  %s  %s  %s  %s\n",
		padRight("period", 12),
		padLeft("opening", 10),
		padLeft("interest", 10),
		padLeft("payments", 10),
		padLeft("closing", 10),
	)
	fmt.Fprintln(w, strings.Repeat("─", 58))

	for _, e := range timeline {
		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			padRight(e.PeriodEnd.Format("2006-01-02"), 12),
			padLeft(e.OpeningBalance.Format(), 10),
			padLeft(e.Interest.Format(), 10),
			padLeft(e.Payments.Format(), 10),
			padLeft(e.ClosingBalance.Format(), 10),
		)
	}
}

func PrintLoanConfirmation(w io.Writer, l *domain.Loan) {
	fmt.Fprintf(w, "%srecorded %s loan: %s %s (%s)%s\n",
		dim, l.Type, l.OriginalAmount.Format(), l.Description, l.CounterpartyName, reset)
}

func PrintSplitResult(w io.Writer, personal *domain.Transaction, loans []*domain.Loan) {
	if personal != nil {
		fmt.Fprintf(w, "%syour share: %s %s%s\n", dim, personal.Amount.Format(), personal.Description, reset)
	}
	for _, l := range loans {
		fmt.Fprintf(w, "%s%s owes you: %s%s\n", dim, l.CounterpartyName, l.OriginalAmount.Format(), reset)
	}
}
