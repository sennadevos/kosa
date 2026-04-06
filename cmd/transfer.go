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

var transferCmd = &cobra.Command{
	Use:   "transfer <amount>",
	Short: "transfer between accounts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		dateStr, _ := cmd.Flags().GetString("date")
		notes, _ := cmd.Flags().GetString("notes")

		if from == "" || to == "" {
			return fmt.Errorf("both --from and --to are required")
		}

		var date time.Time
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
		}

		txn, err := application.Transfer(cmd.Context(), app.TransferInput{
			Amount:      amount,
			FromAccount: from,
			ToAccount:   to,
			Date:        date,
			Notes:       notes,
		})
		if err != nil {
			return err
		}

		output.PrintTransactionConfirmation(os.Stdout, txn)
		return nil
	},
}

func init() {
	transferCmd.Flags().String("from", "", "source account name")
	transferCmd.Flags().String("to", "", "destination account name")
	transferCmd.Flags().String("date", "", "date (YYYY-MM-DD)")
	transferCmd.Flags().String("notes", "", "notes")
	rootCmd.AddCommand(transferCmd)
}
