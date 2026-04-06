package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
)

func PrintBalance(w io.Writer, accountName string, balance domain.Amount, f Format) {
	if f == FormatJSON {
		writeJSON(w, map[string]interface{}{
			"account": accountName,
			"balance": balance.Format(),
		})
		return
	}
	if f == FormatToon {
		fmt.Fprintf(w, "%s|%s\n", accountName, balance.Format())
		return
	}
	fmt.Fprintf(w, "%s: %s\n", accountName, balance.Format())
}

func PrintAllBalances(w io.Writer, balances []app.AccountBalance, f Format) {
	if f == FormatJSON {
		writeJSON(w, balances)
		return
	}

	if len(balances) == 0 {
		fmt.Fprintln(w, "no accounts")
		return
	}

	if f == FormatToon {
		for _, b := range balances {
			fmt.Fprintf(w, "%s|%s\n", b.AccountName, b.Balance.Format())
		}
		return
	}

	total := domain.ZeroAmount()
	fmt.Fprintf(w, "%s  %s\n", padRight("account", 25), padLeft("balance", 12))
	fmt.Fprintln(w, strings.Repeat("─", 40))
	for _, b := range balances {
		fmt.Fprintf(w, "%s  %s\n", padRight(b.AccountName, 25), padLeft(b.Balance.Format(), 12))
		total = total.Add(b.Balance)
	}
	fmt.Fprintln(w, strings.Repeat("─", 40))
	fmt.Fprintf(w, "%s  %s\n", padRight("total", 25), padLeft(total.Format(), 12))
}

func PrintSnapshotConfirmation(w io.Writer, s *domain.BalanceSnapshot) {
	fmt.Fprintf(w, "%srecorded snapshot: %s on %s%s\n",
		dim, s.Balance.Format(), s.Date.Format("2006-01-02"), reset)
}
