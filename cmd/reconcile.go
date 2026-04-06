package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/output"
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "compare actual transactions to recurring rule projections",
	RunE: func(cmd *cobra.Command, args []string) error {
		monthStr, _ := cmd.Flags().GetString("month")

		var year int
		var month time.Month

		if monthStr != "" {
			t, err := time.Parse("2006-01", monthStr)
			if err != nil {
				return fmt.Errorf("month format: YYYY-MM, got %q", monthStr)
			}
			year, month = t.Year(), t.Month()
		} else {
			now := time.Now()
			year, month = now.Year(), now.Month()
		}

		rows, err := application.Reconcile(cmd.Context(), year, month)
		if err != nil {
			return err
		}

		output.PrintReconciliation(os.Stdout, rows, outputFormat())
		return nil
	},
}

func init() {
	reconcileCmd.Flags().String("month", "", "month to reconcile (YYYY-MM, defaults to current)")
	rootCmd.AddCommand(reconcileCmd)
}
