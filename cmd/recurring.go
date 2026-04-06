package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var recurringCmd = &cobra.Command{
	Use:   "recurring",
	Short: "manage recurring rules",
}

var recurringAddCmd = &cobra.Command{
	Use:   "add <name> <amount>",
	Short: "add a recurring rule",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[1])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		txnType, _ := cmd.Flags().GetString("type")
		cat, _ := cmd.Flags().GetString("cat")
		account, _ := cmd.Flags().GetString("account")
		freq, _ := cmd.Flags().GetString("freq")
		day, _ := cmd.Flags().GetInt("day")
		startStr, _ := cmd.Flags().GetString("start")
		notes, _ := cmd.Flags().GetString("notes")

		var startDate time.Time
		if startStr != "" {
			startDate, err = time.Parse("2006-01-02", startStr)
			if err != nil {
				return fmt.Errorf("start date: %w", err)
			}
		}

		rule, err := application.RecurringAdd(cmd.Context(), app.RecurringAddInput{
			Name:       args[0],
			Type:       domain.TransactionType(txnType),
			Amount:     amount,
			Category:   cat,
			Account:    account,
			Frequency:  domain.Frequency(freq),
			DayOfMonth: day,
			StartDate:  startDate,
			Notes:      notes,
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "\033[2mrecorded recurring rule: %s %s %s (%s)\033[0m\n",
			rule.Type, rule.Amount.Format(), rule.Name, rule.Frequency)
		return nil
	},
}

var recurringListCmd = &cobra.Command{
	Use:   "list",
	Short: "list recurring rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := application.RecurringList(cmd.Context())
		if err != nil {
			return err
		}

		f := outputFormat()
		if f == output.FormatJSON {
			output.WriteJSONPublic(os.Stdout, rules)
			return nil
		}

		if len(rules) == 0 {
			fmt.Fprintln(os.Stdout, "no recurring rules")
			return nil
		}

		fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
			padR("id", 10), padR("type", 8), padL("amount", 10),
			padR("name", 20), padR("freq", 10), padR("active", 6))
		fmt.Fprintln(os.Stdout, strings.Repeat("─", 70))

		for _, r := range rules {
			active := "\033[32myes\033[0m"
			if !r.IsActive {
				active = "\033[2mno\033[0m"
			}
			fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
				padR(r.ID, 10), padR(string(r.Type), 8), padL(r.Amount.Format(), 10),
				padR(r.Name, 20), padR(string(r.Frequency), 10), active)
		}
		return nil
	},
}

var recurringPauseCmd = &cobra.Command{
	Use:   "pause <rule-id>",
	Short: "pause a recurring rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := application.RecurringPause(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "\033[2mpaused rule %s\033[0m\n", args[0])
		return nil
	},
}

var recurringResumeCmd = &cobra.Command{
	Use:   "resume <rule-id>",
	Short: "resume a recurring rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := application.RecurringResume(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "\033[2mresumed rule %s\033[0m\n", args[0])
		return nil
	},
}

func padR(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padL(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return strings.Repeat(" ", n-len(s)) + s
}

func init() {
	recurringAddCmd.Flags().String("type", "expense", "transaction type: income/expense")
	recurringAddCmd.Flags().String("cat", "", "category name")
	recurringAddCmd.Flags().String("account", "", "account name")
	recurringAddCmd.Flags().String("freq", "monthly", "frequency: daily/weekly/biweekly/monthly/quarterly/yearly")
	recurringAddCmd.Flags().Int("day", 0, "day of month (for monthly)")
	recurringAddCmd.Flags().String("start", "", "start date (YYYY-MM-DD, defaults to today)")
	recurringAddCmd.Flags().String("notes", "", "notes")

	recurringCmd.AddCommand(recurringAddCmd, recurringListCmd, recurringPauseCmd, recurringResumeCmd)
	rootCmd.AddCommand(recurringCmd)
}
