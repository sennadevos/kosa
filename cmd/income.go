package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var incomeCmd = &cobra.Command{
	Use:     "income <amount> <description>",
	Aliases: []string{"i"},
	Short:   "record income",
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

		txn, err := application.Income(cmd.Context(), app.IncomeInput{
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

func init() {
	incomeCmd.Flags().String("cat", "", "category name")
	incomeCmd.Flags().String("account", "", "account name")
	incomeCmd.Flags().StringSlice("tag", nil, "tag name (repeatable)")
	incomeCmd.Flags().String("date", "", "date (YYYY-MM-DD)")
	incomeCmd.Flags().String("foreign", "", "foreign amount and currency, e.g. '500 USD'")
	incomeCmd.Flags().String("ref", "", "reference number")
	incomeCmd.Flags().String("notes", "", "notes")
	rootCmd.AddCommand(incomeCmd)
}
