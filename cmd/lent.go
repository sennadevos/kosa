package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var lentCmd = &cobra.Command{
	Use:   "lent <amount> <description>",
	Short: "quick: someone owes you money",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}
		to, _ := cmd.Flags().GetString("to")
		if to == "" {
			return fmt.Errorf("--to is required")
		}
		dateStr, _ := cmd.Flags().GetString("date")
		var date time.Time
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
		}

		loan, err := application.Lent(cmd.Context(), amount, args[1], to, date)
		if err != nil {
			return err
		}
		output.PrintLoanConfirmation(os.Stdout, loan)
		return nil
	},
}

func init() {
	lentCmd.Flags().String("to", "", "who owes you")
	lentCmd.Flags().String("date", "", "loan date (YYYY-MM-DD, default today)")
	rootCmd.AddCommand(lentCmd)
}
