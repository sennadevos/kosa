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

var refundCmd = &cobra.Command{
	Use:   "refund <amount> <description>",
	Short: "record a refund",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		cat, _ := cmd.Flags().GetString("cat")
		account, _ := cmd.Flags().GetString("account")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		dateStr, _ := cmd.Flags().GetString("date")
		ofID, _ := cmd.Flags().GetString("of")
		notes, _ := cmd.Flags().GetString("notes")

		var date time.Time
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
		}

		txn, err := application.Refund(cmd.Context(), app.RefundInput{
			Amount:      amount,
			Description: args[1],
			Category:    cat,
			Account:     account,
			Tags:        tags,
			Date:        date,
			RefundOfID:  ofID,
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
	refundCmd.Flags().String("cat", "", "category name")
	refundCmd.Flags().String("account", "", "account name")
	refundCmd.Flags().StringSlice("tag", nil, "tag name (repeatable)")
	refundCmd.Flags().String("date", "", "date (YYYY-MM-DD)")
	refundCmd.Flags().String("of", "", "ID of the original transaction being refunded")
	refundCmd.Flags().String("notes", "", "notes")
	rootCmd.AddCommand(refundCmd)
}
