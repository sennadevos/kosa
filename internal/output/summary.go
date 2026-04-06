package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
)

func PrintSummary(w io.Writer, summaries []app.CategorySummary, f Format) {
	if f == FormatJSON {
		writeJSON(w, summaries)
		return
	}

	if len(summaries) == 0 {
		fmt.Fprintln(w, "no spending in this period")
		return
	}

	if f == FormatToon {
		for _, s := range summaries {
			fmt.Fprintf(w, "%s|%s|%s|%s\n",
				s.CategoryName, s.Expenses.Format(), s.Refunds.Format(), s.Net.Format())
		}
		return
	}

	fmt.Fprintf(w, "%s  %s  %s  %s\n",
		padRight("category", 20),
		padLeft("expenses", 10),
		padLeft("refunds", 10),
		padLeft("net", 10),
	)
	fmt.Fprintln(w, strings.Repeat("─", 55))

	total := domain.ZeroAmount()
	for _, s := range summaries {
		net := s.Net.Format()
		if s.Net.IsPositive() {
			net = red + net + reset
		}
		fmt.Fprintf(w, "%s  %s  %s  %s\n",
			padRight(s.CategoryName, 20),
			padLeft(s.Expenses.Format(), 10),
			padLeft(s.Refunds.Format(), 10),
			padLeft(net, 10),
		)
		total = total.Add(s.Net)
	}

	fmt.Fprintln(w, strings.Repeat("─", 55))
	totalStr := total.Format()
	if total.IsPositive() {
		totalStr = red + totalStr + reset
	}
	fmt.Fprintf(w, "%s  %s  %s  %s\n",
		padRight("total", 20),
		padLeft("", 10),
		padLeft("", 10),
		padLeft(totalStr, 10),
	)
}
