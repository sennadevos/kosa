package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/sennadevos/kosa/internal/app"
)

func PrintReconciliation(w io.Writer, rows []app.ReconcileRow, f Format) {
	if f == FormatJSON {
		writeJSON(w, rows)
		return
	}

	if len(rows) == 0 {
		fmt.Fprintln(w, "no recurring rules to reconcile")
		return
	}

	if f == FormatToon {
		for _, r := range rows {
			fmt.Fprintf(w, "%s|%s|%s|%s|%s\n",
				r.RuleName, r.Expected.Format(), r.Actual.Format(), r.Delta.Format(), r.Status)
		}
		return
	}

	fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
		padRight("rule", 20),
		padLeft("expected", 10),
		padLeft("actual", 10),
		padLeft("delta", 10),
		padRight("status", 10),
	)
	fmt.Fprintln(w, strings.Repeat("─", 65))

	for _, r := range rows {
		status := r.Status
		switch status {
		case "linked":
			if r.Delta.IsZero() {
				status = green + "ok" + reset
			} else {
				status = red + "delta" + reset
			}
		case "missing":
			status = red + "missing" + reset
		}

		deltaStr := r.Delta.Format()
		if !r.Delta.IsZero() {
			if r.Delta.IsPositive() {
				deltaStr = green + "+" + deltaStr + reset
			} else {
				deltaStr = red + deltaStr + reset
			}
		}

		actualStr := r.Actual.Format()
		if r.Status == "missing" {
			actualStr = dim + "-" + reset
			deltaStr = dim + "-" + reset
		}

		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			padRight(r.RuleName, 20),
			padLeft(r.Expected.Format(), 10),
			padLeft(actualStr, 10),
			padLeft(deltaStr, 10),
			status,
		)
	}
}
