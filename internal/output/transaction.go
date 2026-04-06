package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/sennadevos/kosa/internal/domain"
)

func PrintTransaction(w io.Writer, t *domain.Transaction, f Format) {
	PrintTransactions(w, []domain.Transaction{*t}, f)
}

func PrintTransactions(w io.Writer, txns []domain.Transaction, f Format) {
	switch f {
	case FormatJSON:
		writeJSON(w, txns)
	case FormatToon:
		printTransactionsToon(w, txns)
	default:
		printTransactionsTable(w, txns)
	}
}

func printTransactionsTable(w io.Writer, txns []domain.Transaction) {
	if len(txns) == 0 {
		fmt.Fprintln(w, "no transactions")
		return
	}

	// header
	fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
		padRight("date", 10),
		padRight("type", 8),
		padLeft("amount", 10),
		padRight("description", 30),
		padRight("category", 15),
	)
	fmt.Fprintln(w, strings.Repeat("─", 80))

	for _, t := range txns {
		cat := t.CategoryName
		if cat == "" {
			cat = dim + "-" + reset
		}
		amount := colorAmount(string(t.Type), t.Amount.Format())

		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			padRight(t.Date.Format("2006-01-02"), 10),
			padRight(string(t.Type), 8),
			padLeft(amount, 10+len(amount)-len(t.Amount.Format())-1), // account for ANSI codes
			padRight(t.Description, 30),
			cat,
		)
	}
}

func printTransactionsToon(w io.Writer, txns []domain.Transaction) {
	for _, t := range txns {
		parts := []string{
			t.Date.Format("2006-01-02"),
			string(t.Type),
			t.Amount.Format(),
			t.Description,
		}
		if t.CategoryName != "" {
			parts = append(parts, "cat:"+t.CategoryName)
		}
		if t.AccountName != "" {
			parts = append(parts, "acc:"+t.AccountName)
		}
		fmt.Fprintln(w, strings.Join(parts, "|"))
	}
}

func PrintTransactionConfirmation(w io.Writer, t *domain.Transaction) {
	fmt.Fprintf(w, "%srecorded %s: %s %s%s\n",
		dim, t.Type, t.Amount.Format(), t.Description, reset)
}
