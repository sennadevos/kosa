package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/output"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "spending summary by category",
	RunE: func(cmd *cobra.Command, args []string) error {
		monthStr, _ := cmd.Flags().GetString("month")
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")

		var from, to time.Time

		if monthStr != "" {
			if monthStr == "current" || monthStr == "" {
				now := time.Now()
				from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			} else {
				t, err := time.Parse("2006-01", monthStr)
				if err != nil {
					return fmt.Errorf("month format: YYYY-MM, got %q", monthStr)
				}
				from = t
			}
			to = from.AddDate(0, 1, -1)
		} else if fromStr != "" && toStr != "" {
			var err error
			from, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return fmt.Errorf("from date: %w", err)
			}
			to, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("to date: %w", err)
			}
		} else {
			// default to current month
			now := time.Now()
			from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			to = from.AddDate(0, 1, -1)
		}

		summaries, err := application.SpendingSummary(cmd.Context(), from, to)
		if err != nil {
			return err
		}

		output.PrintSummary(os.Stdout, summaries, outputFormat())
		return nil
	},
}

func init() {
	summaryCmd.Flags().String("month", "", "month (YYYY-MM or 'current')")
	summaryCmd.Flags().String("from", "", "start date (YYYY-MM-DD)")
	summaryCmd.Flags().String("to", "", "end date (YYYY-MM-DD)")
	rootCmd.AddCommand(summaryCmd)
}
