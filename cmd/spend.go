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

var spendCmd = &cobra.Command{
	Use:     "spend <amount> <description>",
	Aliases: []string{"s"},
	Short:   "record an expense",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		cat, _ := cmd.Flags().GetString("cat")
		account, _ := cmd.Flags().GetString("account")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		dateStr, _ := cmd.Flags().GetString("date")
		foreignStr, _ := cmd.Flags().GetString("foreign")
		ref, _ := cmd.Flags().GetString("ref")
		notes, _ := cmd.Flags().GetString("notes")

		var date time.Time
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
		}

		var foreignAmount domain.Amount
		var foreignCurrency string
		if foreignStr != "" {
			foreignAmount, foreignCurrency, err = parseForeign(foreignStr)
			if err != nil {
				return err
			}
		}

		txn, err := application.Spend(cmd.Context(), app.SpendInput{
			Amount:          amount,
			Description:     args[1],
			Category:        cat,
			Account:         account,
			Tags:            tags,
			Date:            date,
			ForeignAmount:   foreignAmount,
			ForeignCurrency: foreignCurrency,
			Reference:       ref,
			Notes:           notes,
		})
		if err != nil {
			return err
		}

		output.PrintTransactionConfirmation(os.Stdout, txn)
		return nil
	},
}

func parseForeign(s string) (domain.Amount, string, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return domain.ZeroAmount(), "", fmt.Errorf("--foreign must be '<amount> <currency>', e.g. '500 USD'")
	}
	amount, err := domain.NewAmount(parts[0])
	if err != nil {
		return domain.ZeroAmount(), "", fmt.Errorf("foreign amount: %w", err)
	}
	return amount, strings.ToUpper(parts[1]), nil
}

func init() {
	spendCmd.Flags().String("cat", "", "category name")
	spendCmd.Flags().String("account", "", "account name (defaults to configured default)")
	spendCmd.Flags().StringSlice("tag", nil, "tag name (repeatable)")
	spendCmd.Flags().String("date", "", "date (YYYY-MM-DD, defaults to today)")
	spendCmd.Flags().String("foreign", "", "foreign amount and currency, e.g. '50 GBP'")
	spendCmd.Flags().String("ref", "", "reference number")
	spendCmd.Flags().String("notes", "", "notes")
	rootCmd.AddCommand(spendCmd)
}
