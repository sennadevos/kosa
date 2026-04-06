package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/output"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "show current or projected balance",
	RunE: func(cmd *cobra.Command, args []string) error {
		onStr, _ := cmd.Flags().GetString("on")
		account, _ := cmd.Flags().GetString("account")

		f := outputFormat()

		// specific account with future date
		if onStr != "" && account != "" {
			targetDate, err := time.Parse("2006-01-02", onStr)
			if err != nil {
				return fmt.Errorf("date: %w", err)
			}
			accID, err := application.ResolveAccountID(cmd.Context(), account)
			if err != nil {
				return err
			}
			bal, err := application.ProjectedBalance(cmd.Context(), accID, targetDate)
			if err != nil {
				return err
			}
			output.PrintBalance(os.Stdout, account, bal, f)
			return nil
		}

		// specific account, current balance
		if account != "" {
			accID, err := application.ResolveAccountID(cmd.Context(), account)
			if err != nil {
				return err
			}
			bal, err := application.CurrentBalance(cmd.Context(), accID)
			if err != nil {
				return err
			}
			output.PrintBalance(os.Stdout, account, bal, f)
			return nil
		}

		// all accounts
		balances, err := application.AllBalances(cmd.Context())
		if err != nil {
			return err
		}
		output.PrintAllBalances(os.Stdout, balances, f)
		return nil
	},
}

func init() {
	balanceCmd.Flags().String("on", "", "projected balance on date (YYYY-MM-DD)")
	balanceCmd.Flags().String("account", "", "specific account")
	rootCmd.AddCommand(balanceCmd)
}
